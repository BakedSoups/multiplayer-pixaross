create extension if not exists pgcrypto;

create type public.content_status as enum ('draft', 'published', 'hidden', 'removed');
create type public.content_visibility as enum ('draft', 'public', 'pack_only', 'unlisted');
create type public.submission_status as enum ('submitted', 'in_review', 'changes_requested', 'approved', 'declined');
create type public.report_status as enum ('open', 'reviewed', 'dismissed', 'actioned');

create table public.profiles (
  id uuid primary key references auth.users(id) on delete cascade,
  display_name text not null default 'Creator' check (char_length(display_name) between 1 and 40),
  bio text not null default '' check (char_length(bio) <= 120),
  role text not null default 'creator' check (role in ('creator', 'moderator', 'admin')),
  created_at timestamptz not null default now()
);

create table public.drafts (
  id uuid primary key default gen_random_uuid(),
  owner_id uuid not null references public.profiles(id) on delete cascade,
  local_id text,
  title text not null check (char_length(title) between 1 and 80),
  description text not null default '' check (char_length(description) <= 500),
  tags text[] not null default '{}',
  puzzle jsonb not null,
  playtested boolean not null default false,
  updated_at timestamptz not null default now(),
  unique (owner_id, local_id)
);

create table public.levels (
  id uuid primary key default gen_random_uuid(),
  owner_id uuid not null references public.profiles(id) on delete cascade,
  local_id text,
  title text not null check (char_length(title) between 1 and 80),
  description text not null default '' check (char_length(description) <= 500),
  tags text[] not null default '{}',
  visibility public.content_visibility not null default 'public',
  status public.content_status not null default 'published',
  current_version integer not null default 1 check (current_version > 0),
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now(),
  unique (owner_id, local_id)
);

create table public.level_versions (
  id uuid primary key default gen_random_uuid(),
  level_id uuid not null references public.levels(id) on delete cascade,
  version integer not null check (version > 0),
  title text not null,
  description text not null default '',
  tags text[] not null default '{}',
  puzzle jsonb not null,
  published_at timestamptz not null default now(),
  unique (level_id, version)
);

create table public.packs (
  id uuid primary key default gen_random_uuid(),
  owner_id uuid not null references public.profiles(id) on delete cascade,
  title text not null check (char_length(title) between 1 and 80),
  description text not null default '' check (char_length(description) <= 500),
  tags text[] not null default '{}',
  visibility public.content_visibility not null default 'public',
  status public.content_status not null default 'published',
  current_version integer not null default 1,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

create table public.pack_versions (
  id uuid primary key default gen_random_uuid(),
  pack_id uuid not null references public.packs(id) on delete cascade,
  version integer not null,
  title text not null,
  description text not null default '',
  published_at timestamptz not null default now(),
  unique (pack_id, version)
);

create table public.pack_items (
  pack_version_id uuid not null references public.pack_versions(id) on delete cascade,
  level_version_id uuid not null references public.level_versions(id) on delete restrict,
  position integer not null check (position between 0 and 19),
  primary key (pack_version_id, position),
  unique (pack_version_id, level_version_id)
);

create table public.pack_progress (
  user_id uuid not null references public.profiles(id) on delete cascade,
  pack_id uuid not null references public.packs(id) on delete cascade,
  completed_level_ids uuid[] not null default '{}',
  updated_at timestamptz not null default now(),
  primary key (user_id, pack_id)
);

create table public.official_submissions (
  id uuid primary key default gen_random_uuid(),
  owner_id uuid not null references public.profiles(id) on delete cascade,
  level_id uuid not null references public.levels(id) on delete cascade,
  level_version_id uuid not null references public.level_versions(id) on delete restrict,
  status public.submission_status not null default 'submitted',
  rights_confirmed boolean not null check (rights_confirmed),
  creator_note text not null default '' check (char_length(creator_note) <= 1000),
  reviewer_note text not null default '' check (char_length(reviewer_note) <= 1000),
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

create table public.likes (
  user_id uuid not null references public.profiles(id) on delete cascade,
  level_id uuid not null references public.levels(id) on delete cascade,
  created_at timestamptz not null default now(),
  primary key (user_id, level_id)
);

create table public.play_events (
  id bigint generated always as identity primary key,
  user_id uuid references public.profiles(id) on delete set null,
  level_id uuid not null references public.levels(id) on delete cascade,
  completed boolean not null default false,
  duration_seconds integer check (duration_seconds >= 0),
  created_at timestamptz not null default now()
);

create table public.reports (
  id uuid primary key default gen_random_uuid(),
  reporter_id uuid references public.profiles(id) on delete set null,
  level_id uuid references public.levels(id) on delete cascade,
  pack_id uuid references public.packs(id) on delete cascade,
  reason text not null check (char_length(reason) between 3 and 500),
  status public.report_status not null default 'open',
  created_at timestamptz not null default now(),
  check ((level_id is null) <> (pack_id is null))
);

create table public.notifications (
  id uuid primary key default gen_random_uuid(),
  user_id uuid not null references public.profiles(id) on delete cascade,
  kind text not null,
  body text not null,
  read_at timestamptz,
  created_at timestamptz not null default now()
);

create or replace function public.create_profile_for_user()
returns trigger language plpgsql security definer set search_path = public as $$
begin
  insert into public.profiles (id) values (new.id) on conflict do nothing;
  return new;
end;
$$;
create trigger auth_user_profile after insert on auth.users for each row execute function public.create_profile_for_user();

create or replace function public.reject_version_mutation()
returns trigger language plpgsql as $$ begin raise exception 'published versions are immutable'; end; $$;
create trigger immutable_level_versions before update or delete on public.level_versions for each row execute function public.reject_version_mutation();
create trigger immutable_pack_versions before update or delete on public.pack_versions for each row execute function public.reject_version_mutation();

create or replace function public.publish_level(
  p_local_id text,
  p_title text,
  p_description text,
  p_tags text[],
  p_puzzle jsonb,
  p_submit_official boolean default false,
  p_rights_confirmed boolean default false
) returns jsonb language plpgsql security definer set search_path = public as $$
declare
  uid uuid := auth.uid();
  target_level public.levels;
  target_version public.level_versions;
  next_version integer;
  width integer := (p_puzzle->>'width')::integer;
  height integer := (p_puzzle->>'height')::integer;
	row_text text;
	pixel_row jsonb;
	pixel_value text;
	filled_cells integer := 0;
begin
  if uid is null then raise exception 'Sign in before publishing'; end if;
  if trim(p_title) = '' or char_length(p_title) > 80 then raise exception 'Title must be 1 to 80 characters'; end if;
  if width not in (8, 10, 15, 20) or height not in (8, 10, 15, 20) then raise exception 'Unsupported puzzle dimensions'; end if;
  if jsonb_typeof(p_puzzle->'solution') <> 'array' or jsonb_array_length(p_puzzle->'solution') <> height then raise exception 'Invalid solution rows'; end if;
	for row_text in select jsonb_array_elements_text(p_puzzle->'solution') loop
		if char_length(row_text) <> width or row_text !~ '^[01]+$' then raise exception 'Invalid solution row'; end if;
		filled_cells := filled_cells + char_length(replace(row_text, '0', ''));
	end loop;
	if filled_cells = 0 or filled_cells = width * height then raise exception 'Puzzle must contain both filled and empty cells'; end if;
	if jsonb_typeof(p_puzzle->'skeletonPixels') <> 'array' or jsonb_array_length(p_puzzle->'skeletonPixels') <> height then raise exception 'Invalid Before layer'; end if;
	for pixel_row in select jsonb_array_elements(p_puzzle->'skeletonPixels') loop
		if jsonb_typeof(pixel_row) <> 'array' or jsonb_array_length(pixel_row) <> width then raise exception 'Invalid Before row'; end if;
		for pixel_value in select jsonb_array_elements_text(pixel_row) loop
			if lower(pixel_value) not in ('', 'transparent', '#000000ff') then raise exception 'Before art must be black or transparent'; end if;
		end loop;
	end loop;
	if jsonb_typeof(p_puzzle->'revealPixels') <> 'array' or jsonb_array_length(p_puzzle->'revealPixels') <> height then raise exception 'Invalid After layer'; end if;
	for pixel_row in select jsonb_array_elements(p_puzzle->'revealPixels') loop
		if jsonb_typeof(pixel_row) <> 'array' or jsonb_array_length(pixel_row) <> width then raise exception 'Invalid After row'; end if;
	end loop;
  if p_submit_official and not p_rights_confirmed then raise exception 'Rights confirmation is required'; end if;

  insert into public.levels (owner_id, local_id, title, description, tags)
  values (uid, p_local_id, trim(p_title), coalesce(p_description, ''), coalesce(p_tags, '{}'))
  on conflict (owner_id, local_id) do update set
    title = excluded.title, description = excluded.description, tags = excluded.tags,
    status = 'published', visibility = 'public',
    current_version = public.levels.current_version + 1, updated_at = now()
  returning * into target_level;

  next_version := target_level.current_version;
  insert into public.level_versions (level_id, version, title, description, tags, puzzle)
  values (target_level.id, next_version, target_level.title, target_level.description, target_level.tags, p_puzzle)
  returning * into target_version;

  if p_submit_official then
    insert into public.official_submissions (owner_id, level_id, level_version_id, rights_confirmed)
    values (uid, target_level.id, target_version.id, true);
  end if;
  return jsonb_build_object('levelId', target_level.id, 'levelVersionId', target_version.id, 'version', next_version);
end;
$$;

create or replace function public.save_draft(p_draft jsonb)
returns void language plpgsql security definer set search_path = public as $$
declare
	uid uuid := auth.uid();
	incoming_updated_at timestamptz := coalesce((p_draft->>'updatedAt')::timestamptz, now());
begin
	if uid is null then raise exception 'Sign in before cloud saving'; end if;
	if coalesce(p_draft->>'id', '') = '' then raise exception 'Draft ID is required'; end if;
	if char_length(coalesce(p_draft->>'title', '')) not between 1 and 80 then raise exception 'Invalid title'; end if;
	insert into public.drafts (owner_id, local_id, title, description, tags, puzzle, playtested, updated_at)
	values (
		uid, p_draft->>'id', p_draft->>'title', coalesce(p_draft->>'description', ''),
		coalesce(array(select jsonb_array_elements_text(p_draft->'tags')), '{}'), p_draft->'puzzle',
		coalesce((p_draft->>'playtested')::boolean, false), incoming_updated_at
	)
	on conflict (owner_id, local_id) do update set
		title = excluded.title, description = excluded.description, tags = excluded.tags,
		puzzle = excluded.puzzle, playtested = excluded.playtested, updated_at = excluded.updated_at
	where public.drafts.updated_at <= excluded.updated_at;
end;
$$;

create or replace function public.review_official_submission(p_submission_id uuid, p_status public.submission_status, p_note text default '')
returns void language plpgsql security definer set search_path = public as $$
declare submission public.official_submissions;
begin
  if not exists (select 1 from public.profiles where id = auth.uid() and role in ('moderator', 'admin')) then
    raise exception 'Admin access required';
  end if;
  if p_status not in ('in_review', 'changes_requested', 'approved', 'declined') then raise exception 'Invalid review status'; end if;
  update public.official_submissions set status = p_status, reviewer_note = coalesce(p_note, ''), updated_at = now()
  where id = p_submission_id returning * into submission;
  if submission.id is null then raise exception 'Submission not found'; end if;
  insert into public.notifications (user_id, kind, body)
  values (submission.owner_id, 'official_submission', 'Your main-game submission is now ' || replace(p_status::text, '_', ' ') || '.');
end;
$$;

create or replace function public.publish_pack(p_title text, p_description text, p_level_local_ids text[])
returns jsonb language plpgsql security definer set search_path = public as $$
declare
	uid uuid := auth.uid();
	target_pack public.packs;
	target_pack_version public.pack_versions;
	current_local_id text;
	level_version_id uuid;
	position_index integer := 0;
begin
	if uid is null then raise exception 'Sign in before publishing'; end if;
	if trim(p_title) = '' or char_length(p_title) > 80 then raise exception 'Pack title must be 1 to 80 characters'; end if;
	if coalesce(cardinality(p_level_local_ids), 0) not between 1 and 20 then raise exception 'Packs must contain 1 to 20 levels'; end if;
	if (select count(distinct value) from unnest(p_level_local_ids) value) <> cardinality(p_level_local_ids) then raise exception 'A level can appear only once'; end if;

	insert into public.packs (owner_id, title, description) values (uid, trim(p_title), coalesce(p_description, '')) returning * into target_pack;
	insert into public.pack_versions (pack_id, version, title, description)
	values (target_pack.id, 1, target_pack.title, target_pack.description) returning * into target_pack_version;

	foreach current_local_id in array p_level_local_ids loop
		select lv.id into level_version_id
		from public.levels l join public.level_versions lv on lv.level_id = l.id and lv.version = l.current_version
		where l.owner_id = uid and l.local_id = current_local_id and l.status = 'published';
		if level_version_id is null then raise exception 'Publish every pack level before publishing the pack'; end if;
		insert into public.pack_items (pack_version_id, level_version_id, position)
		values (target_pack_version.id, level_version_id, position_index);
		position_index := position_index + 1;
	end loop;
	return jsonb_build_object('packId', target_pack.id, 'packVersionId', target_pack_version.id);
end;
$$;

grant execute on function public.publish_level(text, text, text, text[], jsonb, boolean, boolean) to authenticated;
grant execute on function public.save_draft(jsonb) to authenticated;
grant execute on function public.publish_pack(text, text, text[]) to authenticated;
grant execute on function public.review_official_submission(uuid, public.submission_status, text) to authenticated;

alter table public.profiles enable row level security;
alter table public.drafts enable row level security;
alter table public.levels enable row level security;
alter table public.level_versions enable row level security;
alter table public.packs enable row level security;
alter table public.pack_versions enable row level security;
alter table public.pack_items enable row level security;
alter table public.pack_progress enable row level security;
alter table public.official_submissions enable row level security;
alter table public.likes enable row level security;
alter table public.play_events enable row level security;
alter table public.reports enable row level security;
alter table public.notifications enable row level security;

create policy profiles_public_read on public.profiles for select using (true);
create policy profile_owner_update on public.profiles for update using (id = auth.uid()) with check (id = auth.uid() and role = (select role from public.profiles where id = auth.uid()));
create policy drafts_owner_all on public.drafts for all using (owner_id = auth.uid()) with check (owner_id = auth.uid());
create policy levels_public_read on public.levels for select using (status = 'published' and visibility in ('public', 'unlisted') or owner_id = auth.uid());
create policy levels_owner_write on public.levels for all using (owner_id = auth.uid()) with check (owner_id = auth.uid());
create policy level_versions_public_read on public.level_versions for select using (exists (select 1 from public.levels l where l.id = level_id and (l.owner_id = auth.uid() or (l.status = 'published' and l.visibility in ('public', 'unlisted')))));
create policy packs_public_read on public.packs for select using (status = 'published' and visibility in ('public', 'unlisted') or owner_id = auth.uid());
create policy packs_owner_write on public.packs for all using (owner_id = auth.uid()) with check (owner_id = auth.uid());
create policy pack_versions_public_read on public.pack_versions for select using (exists (select 1 from public.packs p where p.id = pack_id and (p.owner_id = auth.uid() or p.status = 'published')));
create policy pack_items_public_read on public.pack_items for select using (exists (select 1 from public.pack_versions pv join public.packs p on p.id = pv.pack_id where pv.id = pack_version_id and (p.owner_id = auth.uid() or p.status = 'published')));
create policy progress_owner_all on public.pack_progress for all using (user_id = auth.uid()) with check (user_id = auth.uid());
create policy submissions_owner_read on public.official_submissions for select using (owner_id = auth.uid() or exists (select 1 from public.profiles where id = auth.uid() and role in ('moderator', 'admin')));
create policy likes_public_read on public.likes for select using (true);
create policy likes_owner_write on public.likes for all using (user_id = auth.uid()) with check (user_id = auth.uid());
create policy plays_insert on public.play_events for insert with check (user_id is null or user_id = auth.uid());
create policy reports_owner_insert on public.reports for insert with check (reporter_id is null or reporter_id = auth.uid());
create policy reports_admin_read on public.reports for select using (exists (select 1 from public.profiles where id = auth.uid() and role in ('moderator', 'admin')));
create policy notifications_owner_read on public.notifications for select using (user_id = auth.uid());
create policy notifications_owner_update on public.notifications for update using (user_id = auth.uid()) with check (user_id = auth.uid());

create index levels_browse_idx on public.levels (status, visibility, updated_at desc);
create index packs_browse_idx on public.packs (status, visibility, updated_at desc);
create index submissions_review_idx on public.official_submissions (status, created_at);
