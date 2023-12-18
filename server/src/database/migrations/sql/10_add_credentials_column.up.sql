alter table users
  add column if not exists credentials JSONB;
