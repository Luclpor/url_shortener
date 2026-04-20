alter table url_shortener drop CONSTRAINT if exists original_url_user_id_unique;
ALTER TABLE url_shortener
DROP COLUMN IF EXISTS user_id;

DROP TRIGGER IF EXISTS tiub_service_user_audit ON service_user;

DROP TABLE IF EXISTS service_user;

