# รัน Web Server
run:
	go run cmd/api/main.go

# ⬆️ สร้างตาราง (Up)
migrate-up:
	go run cmd/migrate/main.go -action=up

# ⬇️ ถอยหลังตาราง (Down)
migrate-down:
	go run cmd/migrate/main.go -action=down

# 💣 ล้างบางตารางทั้งหมด (Drop) - เหมาะสำหรับเคลียร์ตอนเทส
migrate-drop:
	go run cmd/migrate/main.go -action=drop