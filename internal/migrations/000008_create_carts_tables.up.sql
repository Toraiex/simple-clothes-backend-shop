-- 1. สร้างตารางตะกร้าสินค้า (Carts)
CREATE TABLE carts (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(user_id) -- 👈 บังคับให้ 1 User มีตะกร้าที่กำลังใช้งานได้แค่ 1 ใบ
);

-- 2. สร้างตารางรายการสินค้าในตะกร้า (Cart Items)
CREATE TABLE cart_items (
    id SERIAL PRIMARY KEY,
    cart_id INT NOT NULL REFERENCES carts(id) ON DELETE CASCADE,
    variant_id INT NOT NULL REFERENCES product_variants(id) ON DELETE CASCADE,
    quantity INT NOT NULL CHECK (quantity > 0), -- 👈 ห้ามใส่จำนวนติดลบหรือ 0
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(cart_id, variant_id) -- 👈 บังคับว่าใน 1 ตะกร้า ห้ามมี Variant ซ้ำกัน (ถ้าซ้ำให้ใช้วิธีบวกเลข quantity เพิ่มเอา)
);