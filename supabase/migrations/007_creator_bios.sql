alter table public.profiles
  add column if not exists bio text not null default '' check (char_length(bio) <= 120);

create or replace function public.save_creator_profile(p_avatar_puzzle jsonb, p_bio text default '')
returns void language plpgsql security definer set search_path = public as $$
begin
  if auth.uid() is null then
    raise exception 'Sign in before saving a profile';
  end if;
  update public.profiles
  set avatar_puzzle = p_avatar_puzzle,
      bio = left(coalesce(p_bio, ''), 120)
  where id = auth.uid();
end;
$$;

do $$
declare
  definition text;
begin
  select pg_get_functiondef('public.browse_creators()'::regprocedure)
  into definition;

  if position('''bio''' in definition) = 0 then
    definition := replace(
      definition,
      '''displayName'', profile.display_name, ''avatarPuzzle''',
      '''displayName'', profile.display_name, ''bio'', profile.bio, ''avatarPuzzle'''
    );
    execute definition;
  end if;
end;
$$;

grant execute on function public.save_creator_profile(jsonb, text) to authenticated;
