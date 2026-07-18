create or replace function public.browse_my_published()
returns jsonb language sql stable security definer set search_path = public as $$
  select coalesce(jsonb_agg(item order by item->>'publishedAt' desc), '[]'::jsonb)
  from (
    select value as item from jsonb_array_elements(public.browse_gallery('art', 'new'))
    where value->>'ownerId' = auth.uid()::text
    union all
    select value as item from jsonb_array_elements(public.browse_gallery('pack', 'new'))
    where value->>'ownerId' = auth.uid()::text
  ) owned;
$$;

create or replace function public.set_profile_promotion(p_kind text, p_content_id uuid)
returns void language plpgsql security definer set search_path = public as $$
begin
  if auth.uid() is null then raise exception 'Sign in to promote work'; end if;
  if p_kind = 'art' then
    if not exists(select 1 from public.levels where id = p_content_id and owner_id = auth.uid() and status = 'published' and visibility = 'public') then raise exception 'Published art not found'; end if;
    delete from public.profile_promotions where owner_id = auth.uid();
    insert into public.profile_promotions(owner_id, level_id) values(auth.uid(), p_content_id);
  elsif p_kind = 'pack' then
    if not exists(select 1 from public.packs where id = p_content_id and owner_id = auth.uid() and status = 'published' and visibility = 'public') then raise exception 'Published pack not found'; end if;
    delete from public.profile_promotions where owner_id = auth.uid();
    insert into public.profile_promotions(owner_id, pack_id) values(auth.uid(), p_content_id);
  else
    raise exception 'Invalid gallery kind';
  end if;
end;
$$;

create or replace function public.unpublish_community_item(p_kind text, p_content_id uuid)
returns void language plpgsql security definer set search_path = public as $$
begin
  if auth.uid() is null then raise exception 'Sign in to manage published work'; end if;
  if p_kind = 'art' then
    update public.levels set status = 'hidden', updated_at = now()
    where id = p_content_id and owner_id = auth.uid() and status = 'published';
  elsif p_kind = 'pack' then
    update public.packs set status = 'hidden', updated_at = now()
    where id = p_content_id and owner_id = auth.uid() and status = 'published';
  else
    raise exception 'Invalid published content kind';
  end if;
  if not found then raise exception 'Published item not found'; end if;
  delete from public.profile_promotions
  where owner_id = auth.uid() and (level_id = p_content_id or pack_id = p_content_id);
end;
$$;

grant execute on function public.browse_my_published() to authenticated;
grant execute on function public.set_profile_promotion(text, uuid) to authenticated;
grant execute on function public.unpublish_community_item(text, uuid) to authenticated;
