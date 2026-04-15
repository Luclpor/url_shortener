create or replace function audit_table()
returns trigger
language plpgsql
as
$$
begin
    if tg_op = 'INSERT' then
        if new.created_at is null then
            new.created_at := now();
end if;

        if new.created_by is null or new.created_by = '' then
            new.created_by := current_user;
end if;

        if new.updated_at is null then
            new.updated_at := new.created_at;
end if;

        if new.updated_by is null or new.updated_by = '' then
            new.updated_by := new.created_by;
end if;

    elsif tg_op = 'UPDATE' then
        new.updated_at := now();

        if new.updated_by is null or new.updated_by = '' then
            new.updated_by := current_user;
end if;
end if;

return new;
end;
$$;

create extension if not exists "uuid-ossp";

create table if not exists url_shortener
(
    id          uuid default public.uuid_generate_v4() primary key,
    correlation_id varchar(256) null,
    original_url    varchar(256) not null,
    short_url   varchar(256) not null,

    created_at  timestamp with time zone not null,
                              created_by  varchar(256) not null,
    updated_at  timestamp with time zone,
                              updated_by  varchar(256)
    );


drop trigger if exists tiub_url_shortener_audit on url_shortener;
create trigger tiub_url_shortener_audit
    before insert or update
                         on url_shortener
                         for each row
                         execute procedure audit_table();