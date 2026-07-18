alter table public.levels add column if not exists preview_pixels jsonb;
alter table public.packs add column if not exists preview_pixels jsonb;

create or replace function public.set_content_preview(p_kind text, p_content_id uuid, p_preview_pixels jsonb)
returns void language plpgsql security definer set search_path = public as $$
begin
  if auth.uid() is null then raise exception 'Sign in before uploading a cover'; end if;
  if jsonb_typeof(p_preview_pixels) <> 'array' then raise exception 'Invalid cover image'; end if;
  if p_kind = 'art' then
    update public.levels set preview_pixels = p_preview_pixels, updated_at = now()
    where id = p_content_id and owner_id = auth.uid();
  elsif p_kind = 'pack' then
    update public.packs set preview_pixels = p_preview_pixels, updated_at = now()
    where id = p_content_id and owner_id = auth.uid();
  else
    raise exception 'Invalid content kind';
  end if;
  if not found then raise exception 'Published content not found'; end if;
end;
$$;

do $$
declare definition text;
begin
  select pg_get_functiondef('public.browse_gallery(text,text)'::regprocedure) into definition;
  if position('previewPixels' in definition) = 0 then
    definition := replace(definition, '''puzzle'', version.puzzle', '''previewPixels'', level.preview_pixels, ''puzzle'', version.puzzle');
    definition := replace(definition, '''levels'', coalesce', '''previewPixels'', pack.preview_pixels, ''levels'', coalesce');
    execute definition;
  end if;
end;
$$;

grant execute on function public.set_content_preview(text, uuid, jsonb) to authenticated;
