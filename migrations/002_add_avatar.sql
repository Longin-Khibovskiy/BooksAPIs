ALTER TABLE users ADD COLUMN IF NOT EXISTS avatar_url TEXT DEFAULT 'https://flagstaffatvclub.com/wp-content/uploads/2018/10/Vacancy-1.jpg';

CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
