do $$
declare
  definition text;
begin
  select pg_get_functiondef(
    'public.publish_level(text,text,text,text[],jsonb,boolean,boolean)'::regprocedure
  ) into definition;

  if position('status = ''published''' in definition) = 0 then
    definition := replace(
      definition,
      'current_version = public.levels.current_version + 1, updated_at = now()',
      'status = ''published'', visibility = ''public'', current_version = public.levels.current_version + 1, updated_at = now()'
    );
    execute definition;
  end if;
end;
$$;
