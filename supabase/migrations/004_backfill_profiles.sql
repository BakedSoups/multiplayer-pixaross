insert into public.profiles (id, display_name)
select
  auth_user.id,
  left(coalesce(
    nullif(auth_user.raw_user_meta_data->>'full_name', ''),
    nullif(split_part(auth_user.email, '@', 1), ''),
    'Creator'
  ), 40)
from auth.users auth_user
on conflict (id) do nothing;
