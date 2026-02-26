ALTER TABLE categories DROP CONSTRAINT categories_type_check;
ALTER TABLE categories ALTER COLUMN type DROP NOT NULL;
ALTER TABLE categories DROP CONSTRAINT categories_name_type_key;
ALTER TABLE categories ADD CONSTRAINT categories_name_key UNIQUE (name);
