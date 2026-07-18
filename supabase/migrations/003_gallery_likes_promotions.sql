create table if not exists public.pack_likes (
  user_id uuid not null references public.profiles(id) on delete cascade,
  pack_id uuid not null references public.packs(id) on delete cascade,
  created_at timestamptz not null default now(),
  primary key (user_id, pack_id)
);

create table if not exists public.profile_promotions (
  owner_id uuid primary key references public.profiles(id) on delete cascade,
  level_id uuid references public.levels(id) on delete cascade,
  pack_id uuid references public.packs(id) on delete cascade,
  updated_at timestamptz not null default now(),
  check ((level_id is null) <> (pack_id is null))
);

alter table public.pack_likes enable row level security;
alter table public.profile_promotions enable row level security;

create policy pack_likes_public_read on public.pack_likes for select using (true);
create policy pack_likes_owner_write on public.pack_likes for all
  using (user_id = auth.uid()) with check (user_id = auth.uid());
create policy promotions_public_read on public.profile_promotions for select using (true);
create policy promotions_owner_write on public.profile_promotions for all
  using (owner_id = auth.uid()) with check (owner_id = auth.uid());

create or replace function public.browse_gallery(p_kind text default 'art', p_sort text default 'new')
returns jsonb language plpgsql stable security definer set search_path = public as $$
declare result jsonb;
begin
  if p_kind not in ('art', 'pack') then raise exception 'Invalid gallery kind'; end if;
  if p_sort not in ('new', 'top') then raise exception 'Invalid gallery sort'; end if;

  if p_kind = 'art' then
    select coalesce(jsonb_agg(item order by
      case when p_sort = 'top' then (item->>'likes')::integer end desc,
      (item->>'publishedAt')::timestamptz desc), '[]'::jsonb) into result
    from (
      select jsonb_build_object(
        'kind', 'art', 'id', level.id, 'ownerId', level.owner_id,
        'creatorName', profile.display_name, 'avatarPuzzle', profile.avatar_puzzle, 'title', version.title,
        'description', version.description, 'likes', (select count(*) from public.likes where level_id = level.id),
        'liked', exists(select 1 from public.likes where level_id = level.id and user_id = auth.uid()),
        'owned', level.owner_id = auth.uid(),
        'promoted', exists(select 1 from public.profile_promotions where owner_id = level.owner_id and level_id = level.id),
        'puzzle', version.puzzle, 'publishedAt', version.published_at
      ) item
      from public.levels level
      join public.profiles profile on profile.id = level.owner_id
      join public.level_versions version on version.level_id = level.id and version.version = level.current_version
      where level.status = 'published' and level.visibility = 'public'
      limit 200
    ) gallery;
  else
    select coalesce(jsonb_agg(item order by
      case when p_sort = 'top' then (item->>'likes')::integer end desc,
      (item->>'publishedAt')::timestamptz desc), '[]'::jsonb) into result
    from (
      select jsonb_build_object(
        'kind', 'pack', 'id', pack.id, 'ownerId', pack.owner_id,
        'creatorName', profile.display_name, 'avatarPuzzle', profile.avatar_puzzle, 'title', pack_version.title,
        'description', pack_version.description, 'likes', (select count(*) from public.pack_likes where pack_id = pack.id),
        'liked', exists(select 1 from public.pack_likes where pack_id = pack.id and user_id = auth.uid()),
        'owned', pack.owner_id = auth.uid(),
        'promoted', exists(select 1 from public.profile_promotions where owner_id = pack.owner_id and pack_id = pack.id),
        'levels', coalesce((select jsonb_agg(jsonb_build_object(
          'id', level_version.id, 'levelId', level.id, 'version', level_version.version,
          'title', level_version.title, 'description', level_version.description,
          'tags', level_version.tags, 'puzzle', level_version.puzzle,
          'publishedAt', level_version.published_at
        ) order by pack_item.position)
        from public.pack_items pack_item
        join public.level_versions level_version on level_version.id = pack_item.level_version_id
        join public.levels level on level.id = level_version.level_id
        where pack_item.pack_version_id = pack_version.id), '[]'::jsonb),
        'publishedAt', pack_version.published_at
      ) item
      from public.packs pack
      join public.profiles profile on profile.id = pack.owner_id
      join public.pack_versions pack_version on pack_version.pack_id = pack.id and pack_version.version = pack.current_version
      where pack.status = 'published' and pack.visibility = 'public'
      limit 200
    ) gallery;
  end if;
  return result;
end;
$$;

create or replace function public.toggle_gallery_like(p_kind text, p_content_id uuid)
returns boolean language plpgsql security definer set search_path = public as $$
declare now_liked boolean;
begin
  if auth.uid() is null then raise exception 'Sign in to like published work'; end if;
  if p_kind = 'art' then
    if not exists(select 1 from public.levels where id = p_content_id and status = 'published' and visibility = 'public') then raise exception 'Art not found'; end if;
    delete from public.likes where user_id = auth.uid() and level_id = p_content_id;
    if found then return false; end if;
    insert into public.likes(user_id, level_id) values(auth.uid(), p_content_id);
  elsif p_kind = 'pack' then
    if not exists(select 1 from public.packs where id = p_content_id and status = 'published' and visibility = 'public') then raise exception 'Pack not found'; end if;
    delete from public.pack_likes where user_id = auth.uid() and pack_id = p_content_id;
    if found then return false; end if;
    insert into public.pack_likes(user_id, pack_id) values(auth.uid(), p_content_id);
  else
    raise exception 'Invalid gallery kind';
  end if;
  return true;
end;
$$;

create or replace function public.set_profile_promotion(p_kind text, p_content_id uuid)
returns void language plpgsql security definer set search_path = public as $$
begin
  if auth.uid() is null then raise exception 'Sign in to promote work'; end if;
  if p_kind = 'art' then
    if not exists(select 1 from public.levels where id = p_content_id and owner_id = auth.uid() and status = 'published' and visibility = 'public') then raise exception 'Published art not found'; end if;
    insert into public.profile_promotions(owner_id, level_id, pack_id, updated_at)
    values(auth.uid(), p_content_id, null, now()) on conflict(owner_id) do update
    set level_id = excluded.level_id, pack_id = null, updated_at = now();
  elsif p_kind = 'pack' then
    if not exists(select 1 from public.packs where id = p_content_id and owner_id = auth.uid() and status = 'published' and visibility = 'public') then raise exception 'Published pack not found'; end if;
    insert into public.profile_promotions(owner_id, level_id, pack_id, updated_at)
    values(auth.uid(), null, p_content_id, now()) on conflict(owner_id) do update
    set level_id = null, pack_id = excluded.pack_id, updated_at = now();
  else
    raise exception 'Invalid gallery kind';
  end if;
end;
$$;

create or replace function public.browse_creators()
returns jsonb language sql stable security definer set search_path = public as $$
  select coalesce(jsonb_agg(creator order by creator->>'displayName'), '[]'::jsonb)
  from (
    select jsonb_build_object(
      'id', profile.id, 'displayName', profile.display_name, 'bio', profile.bio, 'avatarPuzzle', profile.avatar_puzzle,
      'featured', coalesce((select jsonb_agg(promoted_item.featured) from (
        select jsonb_build_object(
          'kind', 'art', 'id', level.id, 'ownerId', level.owner_id, 'creatorName', profile.display_name,
          'title', version.title, 'likes', (select count(*) from public.likes where level_id = level.id),
          'puzzle', version.puzzle, 'promoted', true, 'publishedAt', version.published_at
        ) featured from public.profile_promotions promotion
        join public.levels level on level.id = promotion.level_id
        join public.level_versions version on version.level_id = level.id and version.version = level.current_version
        where promotion.owner_id = profile.id
        union all
        select jsonb_build_object(
          'kind', 'pack', 'id', pack.id, 'ownerId', pack.owner_id, 'creatorName', profile.display_name,
          'title', pack_version.title, 'likes', (select count(*) from public.pack_likes where pack_id = pack.id),
          'levels', coalesce((select jsonb_agg(jsonb_build_object('id', lv.id, 'levelId', l.id, 'version', lv.version, 'title', lv.title, 'puzzle', lv.puzzle, 'publishedAt', lv.published_at) order by pi.position)
            from public.pack_items pi join public.level_versions lv on lv.id = pi.level_version_id join public.levels l on l.id = lv.level_id where pi.pack_version_id = pack_version.id), '[]'::jsonb),
          'promoted', true, 'publishedAt', pack_version.published_at
        ) featured from public.profile_promotions promotion
        join public.packs pack on pack.id = promotion.pack_id
        join public.pack_versions pack_version on pack_version.pack_id = pack.id and pack_version.version = pack.current_version
        where promotion.owner_id = profile.id
      ) promoted_item), '[]'::jsonb),
      'levels', coalesce((select jsonb_agg(jsonb_build_object(
        'id', version.id, 'levelId', level.id, 'version', version.version, 'title', version.title,
        'description', version.description, 'tags', version.tags, 'puzzle', version.puzzle, 'publishedAt', version.published_at
      ) order by version.published_at desc)
      from public.levels level join public.level_versions version on version.level_id = level.id and version.version = level.current_version
      where level.owner_id = profile.id and level.status = 'published' and level.visibility = 'public'), '[]'::jsonb)
    ) creator from public.profiles profile order by profile.created_at desc limit 100
  ) creators;
$$;

grant execute on function public.browse_gallery(text, text) to anon, authenticated;
grant execute on function public.toggle_gallery_like(text, uuid) to authenticated;
grant execute on function public.set_profile_promotion(text, uuid) to authenticated;
grant execute on function public.browse_creators() to anon, authenticated;

create index if not exists pack_likes_pack_idx on public.pack_likes(pack_id);
