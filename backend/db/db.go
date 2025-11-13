package db

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

var Pool *pgxpool.Pool

func InitDB() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://postgres:1@localhost:5432/ShoesStoreDB"
	}

	var err error
	Pool, err = pgxpool.New(context.Background(), dsn)
	if err != nil {
		log.Fatalf("❌ Не удалось подключиться к БД: %v", err)
	}

	if err = Pool.Ping(context.Background()); err != nil {
		log.Fatalf("❌ Ошибка при проверке подключения: %v", err)
	}

	fmt.Println("✅ Подключение к БД установлено")
}
