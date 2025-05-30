package db

import (
	"context"
	"log"
	"os"

	"github.com/jackc/pgx/v4/pgxpool"
)

var DB *pgxpool.Pool
var Ctx = context.Background()

func InitPostgres() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:123@localhost:5432/leetcode?sslmode=disable"
	}

	var err error
	DB, err = pgxpool.Connect(Ctx, dbURL)
	if err != nil {
		log.Fatalf("❌ 無法連接 PostgreSQL: %v", err)
	}

	_, err = DB.Exec(Ctx, `
		CREATE TABLE IF NOT EXISTS leetcode_questions (
			slug TEXT PRIMARY KEY,
			title TEXT,
			difficulty TEXT,
			rating INTEGER,
			tags TEXT[],
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`)
	if err != nil {
		log.Fatalf("建立資料表失敗: %v", err)
	}
}

func SaveQuestionToDB(slug, title, difficulty string, rating int, tags []string) {
	_, err := DB.Exec(Ctx, `
		INSERT INTO leetcode_questions (slug, title, difficulty, rating, tags, updated_at)
		VALUES ($1, $2, $3, $4, $5, CURRENT_TIMESTAMP)
		ON CONFLICT (slug) DO UPDATE SET
			title = EXCLUDED.title,
			difficulty = EXCLUDED.difficulty,
			rating = EXCLUDED.rating,
			tags = EXCLUDED.tags,
			updated_at = CURRENT_TIMESTAMP
	`, slug, title, difficulty, rating, tags)
	if err != nil {
		log.Printf("寫入 PostgreSQL 錯誤: %v", err)
	} else {
		log.Printf("已寫入 PostgreSQL：%s (%d)", slug, rating)
	}
}
