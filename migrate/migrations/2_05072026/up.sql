ALTER TABLE users 
ADD COLUMN if not exists created_at TIMESTAMPTZ DEFAULT NOW();