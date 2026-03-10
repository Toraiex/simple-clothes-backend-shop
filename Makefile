# รัน Web Server
run:
	go run cmd/api/main.go
	
	go run cmd/seed/main.go

# ⬆️ สร้างตาราง (Up)
migrate-up:
	go run cmd/migrate/main.go -action=up
	migrate -path internal/migrations -database "postgresql://postgres:1234@localhost:5432/postgres?sslmode=disable" up

# 💣 ล้างบางตารางทั้งหมด (Drop) - เหมาะสำหรับเคลียร์ตอนเทส
migrate-drop:
	go run cmd/migrate/main.go -action=drop
	
	
	migrate -path internal/migrations -database "postgresql://postgres:1234@localhost:5432/postgres?sslmode=disable" down



	migrate -path internal/migrations -database "postgresql://postgres:1234@localhost:5432/postgres?sslmode=disable" down 1


	