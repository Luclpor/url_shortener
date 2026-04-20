ALTER TABLE url_shortener
DROP COLUMN IF EXISTS user_id;

DROP TRIGGER IF EXISTS tiub_service_user_audit ON service_user;

DROP TABLE IF EXISTS service_user;