-- 1. ลบคอลัมน์ product_id อันเก่าทิ้ง
ALTER TABLE order_items DROP COLUMN product_id;

-- 2. เพิ่มคอลัมน์ variant_id เข้าไปแทน และผูก Foreign Key
ALTER TABLE order_items ADD COLUMN variant_id INT REFERENCES product_variants(id) ON DELETE SET NULL;