# รัน Web Server
run:
	go run cmd/api/main.go
	go run cmd/seed/main.go

# ⬆️ สร้างตาราง (Up)
migrate-up:
	go run cmd/migrate/main.go -action=up

# 💣 ล้างบางตารางทั้งหมด (Drop) - เหมาะสำหรับเคลียร์ตอนเทส
migrate-drop:
	go run cmd/migrate/main.go -action=drop