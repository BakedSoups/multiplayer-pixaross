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

grant execute on function public.browse_content_chat(text, uuid) to anon, authenticated;
