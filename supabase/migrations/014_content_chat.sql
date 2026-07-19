create table if not exists public.content_chat_messages (
  id uuid primary key default gen_random_uuid(),
  author_id uuid not null references public.profiles(id) on delete cascade,
  level_id uuid references public.levels(id) on delete cascade,
  pack_id uuid references public.packs(id) on delete cascade,
  message_body text not null check (char_length(trim(message_body)) between 1 and 220),
  created_at timestamptz not null default now(),
  check ((level_id is null) <> (pack_id is null))
);

alter table public.content_chat_messages
  add column if not exists message_body text;

do $$
begin
  if exists (
    select 1
    from information_schema.columns
    where table_schema = 'public'
      and table_name = 'content_chat_messages'
      and column_name = 'body'
  ) then
    execute 'update public.content_chat_messages set message_body = coalesce(message_body, body) where message_body is null';
  end if;
end;
$$;

alter table public.content_chat_messages
  alter column message_body set not null;

alter table public.content_chat_messages
  drop constraint if exists content_chat_messages_body_check;

alter table public.content_chat_messages
  drop constraint if exists content_chat_messages_message_body_check;

alter table public.content_chat_messages
  add constraint content_chat_messages_message_body_check check (char_length(trim(message_body)) between 1 and 220);

alter table public.content_chat_messages enable row level security;

create policy chat_public_read on public.content_chat_messages for select using (true);
create policy chat_author_insert on public.content_chat_messages for insert with check (author_id = auth.uid());

create or replace function public.browse_content_chat(p_kind text, p_content_id uuid)
returns jsonb language sql stable security definer set search_path = public as $$
  select coalesce(jsonb_agg(jsonb_build_object(
    'id', message.id,
    'authorId', message.author_id,
    'authorName', profile.display_name,
    'avatarPuzzle', profile.avatar_puzzle,
    'body', message.message_body,
    'createdAt', message.created_at
  ) order by message.created_at), '[]'::jsonb)
  from (
    select * from public.content_chat_messages
    where (p_kind = 'art' and level_id = p_content_id)
       or (p_kind = 'pack' and pack_id = p_content_id)
    order by created_at desc
    limit 40
  ) message
  join public.profiles profile on profile.id = message.author_id;
$$;

create or replace function public.post_content_chat(p_kind text, p_content_id uuid, p_body text)
returns jsonb language plpgsql security definer set search_path = public as $$
declare
  inserted public.content_chat_messages;
begin
  if auth.uid() is null then raise exception 'Sign in to chat'; end if;
  if char_length(trim(coalesce(p_body, ''))) not between 1 and 220 then raise exception 'Message must be 1 to 220 characters'; end if;
  if p_kind = 'art' then
    if not exists(select 1 from public.levels where id = p_content_id and status = 'published' and visibility = 'public') then
      raise exception 'Art not found';
    end if;
    insert into public.content_chat_messages(author_id, level_id, message_body)
    values(auth.uid(), p_content_id, trim(p_body)) returning * into inserted;
  elsif p_kind = 'pack' then
    if not exists(select 1 from public.packs where id = p_content_id and status = 'published' and visibility = 'public') then
      raise exception 'Pack not found';
    end if;
    insert into public.content_chat_messages(author_id, pack_id, message_body)
    values(auth.uid(), p_content_id, trim(p_body)) returning * into inserted;
  else
    raise exception 'Invalid chat kind';
  end if;
  return jsonb_build_object('id', inserted.id);
end;
$$;

grant execute on function public.browse_content_chat(text, uuid) to anon, authenticated;
grant execute on function public.post_content_chat(text, uuid, text) to authenticated;

create index if not exists content_chat_level_created_idx on public.content_chat_messages(level_id, created_at desc);
create index if not exists content_chat_pack_created_idx on public.content_chat_messages(pack_id, created_at desc);
