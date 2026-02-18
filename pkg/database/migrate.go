package database

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	_ "github.com/lib/pq"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// 1. สร้าง struct สำหรับทำ Logger
type migrationLogger struct{}

// 2. ปรับแต่งฟังก์ชัน Printf เพื่อให้แสดงเลขแบบ 000001
func (l *migrationLogger) Printf(format string, v ...interface{}) {
	// นำ format และค่าต่างๆ มารวมเป็นข้อความ (String) ก่อน
	msg := fmt.Sprintf(format, v...)

	// รูปแบบดั้งเดิมของ migrate จะเป็น "1/u name (time)" หรือ "1/d name (time)"
	// เราจะหั่นข้อความด้วย "/" เป็น 2 ท่อน เพื่อเอาตัวเลขข้างหน้ามาจัดรูปแบบใหม่
	parts := strings.SplitN(msg, "/", 2)

	if len(parts) == 2 {
		var version int
		// แปลงข้อความท่อนแรกให้เป็นตัวเลข
		if _, err := fmt.Sscanf(parts[0], "%d", &version); err == nil {
			// ประกอบข้อความกลับเข้าไปใหม่ โดยบังคับให้ตัวเลขมี 6 หลัก (%06d)
			msg = fmt.Sprintf("%06d/%s", version, parts[1])
		}
	}

	// ใช้ Print ธรรมดา (ไม่ใช้ Printf) เพราะใน msg ของ migrate มีการขึ้นบรรทัดใหม่ (\n) ติดมาอยู่แล้ว
	fmt.Print("📜 " + msg)
}

// 3. สร้างฟังก์ชัน Verbose เพื่อเปิดโหมดแสดงรายละเอียด
func (l *migrationLogger) Verbose() bool {
	return true
}

func RunMigrations(db *sql.DB, action string) {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.Fatalf("could not start sql migration: %v", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://./internal/migrations", // ชี้ไปที่โฟลเดอร์เก็บไฟล์ .sql
		"postgres",
		driver,
	)
	if err != nil {
		log.Fatalf("migration failed: %v", err)
	}

	m.Log = &migrationLogger{}

	// ✅ ใช้ Switch-Case เพื่อเลือกว่าจะทำอะไรตามค่า action ที่ส่งมา
	switch action {
	case "up":
		err = m.Up()
		if err != nil && err != migrate.ErrNoChange {
			log.Fatalf("failed to run migrate up: %v", err)
		}
		if err == migrate.ErrNoChange {
			fmt.Println("✅ No new migrations to apply.")
		} else {
			fmt.Println("✅ All migrations applied successfully.")
		}

	case "down":
		err = m.Down()
		if err != nil && err != migrate.ErrNoChange {
			log.Fatalf("failed to run migrate down: %v", err)
		}
		if err == migrate.ErrNoChange {
			fmt.Println("✅ No migrations to rollback.")
		} else {
			fmt.Println("✅ Rolled back all migrations successfully.")
		}

	case "drop":
		err = m.Drop()
		if err != nil {
			log.Fatalf("failed to drop database schema: %v", err)
		}
		fmt.Println("💣 Dropped all migrations and tables successfully.")

	default:
		log.Fatalf("unknown migration action: %s", action)
	}
}
