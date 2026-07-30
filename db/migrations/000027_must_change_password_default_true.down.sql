ALTER TABLE users ALTER COLUMN must_change_password DROP NOT NULL;
ALTER TABLE users ALTER COLUMN must_change_password SET DEFAULT false;
