create or replace function public.save_creator_profile(p_avatar_puzzle jsonb, p_bio text default '', p_display_name text default 'Creator')
returns void language plpgsql security definer set search_path = public as $$
begin
  if auth.uid() is null then raise exception 'Sign in before saving a profile'; end if;
  if trim(coalesce(p_display_name, '')) = '' then raise exception 'Name is required'; end if;
  update public.profiles
  set avatar_puzzle = p_avatar_puzzle,
      bio = left(coalesce(p_bio, ''), 120),
      display_name = left(trim(p_display_name), 40)
  where id = auth.uid();
end;
$$;

grant execute on function public.save_creator_profile(jsonb, text, text) to authenticated;
