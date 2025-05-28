package mq

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"github.com/streadway/amqp"
)

type QuestionTask struct {
	Slug       string   `json:"slug"`
	Title      string   `json:"title"`
	Difficulty string   `json:"difficulty"`
	Rating     int      `json:"rating"`
	Tags       []string `json:"tags"`
}

var channel *amqp.Channel
var queueName = "question_tasks"

// 初始化 RabbitMQ 並建立 channel 與 queue
func InitRabbitMQ() {
	rabbitURL := os.Getenv("RABBITMQ_URL")
	if rabbitURL == "" {
		rabbitURL = "amqp://guest:guest@localhost:5672/"
	}

	conn, err := amqp.Dial(rabbitURL)
	if err != nil {
		log.Fatalf("❌ 無法連接 RabbitMQ: %v", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("❌ 無法建立 RabbitMQ channel: %v", err)
	}
	channel = ch

	_, err = ch.QueueDeclare(queueName, true, false, false, false, nil)
	if err != nil {
		log.Fatalf("❌ 無法宣告 queue: %v", err)
	}

	log.Println("✅ RabbitMQ 初始化完成")
}

// 發送任務到 RabbitMQ queue
func PublishQuestion(task QuestionTask) {
	body, err := json.Marshal(task)
	if err != nil {
		log.Printf("❌ JSON 編碼失敗: %v", err)
		return
	}

	err = channel.Publish(
		"",        // exchange
		queueName, // routing key (queue)
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
	if err != nil {
		log.Printf("❌ 發送任務失敗: %v", err)
	} else {
		log.Printf("📤 已發送任務至 RabbitMQ：%s", task.Slug)
	}
}

// 啟動 RabbitMQ worker 以處理 queue 中的任務
func StartWorker(ctx context.Context, handler func(task QuestionTask)) {
	rabbitURL := os.Getenv("RABBITMQ_URL")
	if rabbitURL == "" {
		rabbitURL = "amqp://guest:guest@localhost:5672/"
	}

	conn, err := amqp.Dial(rabbitURL)
	if err != nil {
		log.Fatalf("❌ 無法連接 RabbitMQ: %v", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("❌ 無法建立 channel: %v", err)
	}

	_, err = ch.QueueDeclare(queueName, true, false, false, false, nil)
	if err != nil {
		log.Fatalf("❌ 無法宣告 queue: %v", err)
	}

	msgs, err := ch.Consume(queueName, "", true, false, false, false, nil)
	if err != nil {
		log.Fatalf("❌ 無法註冊 consumer: %v", err)
	}

	log.Println("🚀 RabbitMQ Worker 已啟動，等待任務中...")
	for {
		select {
		case msg := <-msgs:
			var task QuestionTask
			if err := json.Unmarshal(msg.Body, &task); err != nil {
				log.Printf("❌ 任務解析錯誤: %v", err)
				continue
			}
			handler(task)
		case <-ctx.Done():
			log.Println("🛑 Worker 停止")
			_ = ch.Close()
			_ = conn.Close()
			return
		}
	}
}
