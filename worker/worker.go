package main

import (
	"context"
	"goprojects/db"
	"goprojects/mq"
)

func main() {
	ctx := context.Background()

	// 初始化資料庫與 RabbitMQ
	// db.InitPostgres()
	mq.InitRabbitMQ()

	// 啟動 Worker
	mq.StartWorker(ctx, func(task mq.QuestionTask) {
		db.SaveQuestionToDB(task.Slug, task.Title, task.Difficulty, task.Rating, task.Tags)
	})
}
