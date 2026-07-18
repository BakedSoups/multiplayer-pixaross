alter table public.profiles
  add column if not exists avatar_puzzle jsonb;

create or replace function public.create_profile_for_user()
returns trigger language plpgsql security definer set search_path = public as $$
begin
  insert into public.profiles (id, display_name)
  values (
    new.id,
    left(coalesce(nullif(new.raw_user_meta_data->>'full_name', ''), nullif(split_part(new.email, '@', 1), ''), 'Creator'), 40)
  )
  on conflict do nothing;
  return new;
end;
$$;

update public.profiles profile
set display_name = left(coalesce(nullif(user_row.raw_user_meta_data->>'full_name', ''), nullif(split_part(user_row.email, '@', 1), ''), 'Creator'), 40)
from auth.users user_row
where profile.id = user_row.id and profile.display_name = 'Creator';

create or replace function public.save_creator_profile(p_avatar_puzzle jsonb, p_bio text default '', p_display_name text default 'Creator')
returns void language plpgsql security definer set search_path = public as $$
begin
  if auth.uid() is null then
    raise exception 'Sign in before saving a profile';
  end if;
  update public.profiles
  set avatar_puzzle = p_avatar_puzzle, bio = left(coalesce(p_bio, ''), 120), display_name = left(trim(p_display_name), 40)
  where id = auth.uid();
end;
$$;

create or replace function public.browse_creators()
returns jsonb language sql stable security definer set search_path = public as $$
  select coalesce(jsonb_agg(creator order by creator->>'displayName'), '[]'::jsonb)
  from (
    select jsonb_build_object(
      'id', profile.id,
      'displayName', profile.display_name,
      'bio', profile.bio,
      'avatarPuzzle', profile.avatar_puzzle,
      'levels', coalesce((
        select jsonb_agg(jsonb_build_object(
          'id', version.id,
          'levelId', level.id,
          'version', version.version,
          'title', version.title,
          'description', version.description,
          'tags', version.tags,
          'puzzle', version.puzzle,
          'publishedAt', version.published_at
        ) order by version.published_at desc)
        from public.levels level
        join public.level_versions version
          on version.level_id = level.id and version.version = level.current_version
        where level.owner_id = profile.id
          and level.status = 'published'
          and level.visibility = 'public'
      ), '[]'::jsonb)
    ) as creator
    from public.profiles profile
    where exists (
      select 1 from public.levels level
      where level.owner_id = profile.id
        and level.status = 'published'
        and level.visibility = 'public'
    )
    order by profile.created_at desc
    limit 50
  ) creators;
$$;

grant execute on function public.save_creator_profile(jsonb, text, text) to authenticated;
grant execute on function public.browse_creators() to anon, authenticated;
