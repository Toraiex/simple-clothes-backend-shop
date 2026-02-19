CREATE TABLE categories (
    id SERIAL PRIMARY KEY,
    name TEXT UNIQUE NOT NULL,
   created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP, -- 👈 ตรงนี้สำคัญมาก
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP  -- 👈 ตรงนี้ด้วย
);
