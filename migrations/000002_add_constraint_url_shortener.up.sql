alter table url_shortener drop  constraint if exists original_url_unique;
alter table url_shortener add  CONSTRAINT original_url_unique UNIQUE (original_url);