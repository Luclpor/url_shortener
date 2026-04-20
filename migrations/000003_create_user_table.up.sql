create table if not exists service_user
(
    id          uuid default public.uuid_generate_v4() primary key,

    created_at  timestamp with time zone not null,
                              created_by  varchar(256) not null,
    updated_at  timestamp with time zone,
                              updated_by  varchar(256)
    );

drop trigger if exists tiub_service_user_audit on service_user;
create trigger tiub_service_user_audit
    before insert or update
                         on service_user
                         for each row
                         execute procedure audit_table();

alter table url_shortener add column user_id uuid null;

alter table url_shortener drop constraint if exists original_url_unique;