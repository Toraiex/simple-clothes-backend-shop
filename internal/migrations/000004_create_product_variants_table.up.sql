CREATE TABLE product_variants (
    id SERIAL PRIMARY KEY,
    product_id INT REFERENCES products(id) ON DELETE CASCADE,
    sku VARCHAR(100) UNIQUE NOT NULL,    -- 👈 เพิ่มรหัส SKU (ห้ามซ้ำ)
    price NUMERIC(10, 2) NOT NULL,
    stock INT DEFAULT 0,
    attributes JSONB DEFAULT '{}'::jsonb, -- 👈 เก็บ Size, Color หรืออื่นๆ ในนี้
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);