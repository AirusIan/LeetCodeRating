package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"goprojects/mq"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

var (
	rdb *redis.Client
	ctx = context.Background()
)

type Tag struct {
	Name string `json:"name"`
}

func main() {
	// 初始化 Redis 連線
	// rdb = redis.NewClient(&redis.Options{
	// 	Addr:     "127.0.0.1:6379", // 明確使用 IPv4 避免連線錯誤
	// 	Password: "",
	// 	DB:       0,
	// })
	rdb = redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_URL"), // 從環境變數讀取
		Password: "",
		DB:       0,
	})

	// 測試 Redis 是否正常運作
	if _, err := rdb.Ping(ctx).Result(); err != nil {
		log.Fatalf("無法連線到 Redis: %v", err)
	}

	// 建立 Gin 路由
	r := gin.Default()

	mq.InitRabbitMQ()

	// 提供靜態網頁首頁（查詢頁面）
	r.StaticFile("/", "./index.html")

	// Rating 查詢 API
	r.GET("/api/question/:slug", handleQueryRating)

	// 啟動 Server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("伺服器啟動於 http://localhost:%s", port)
	r.Run(":" + port)
}

func handleQueryRating(c *gin.Context) {
	slug := c.Param("slug")
	redisKey := fmt.Sprintf("rating:%s", slug)

	// 1. 嘗試從 Redis 快取讀取
	rating, err := rdb.Get(ctx, redisKey).Result()
	if err == nil && rating != "" {
		c.JSON(http.StatusOK, gin.H{"rating": rating})
		return
	}

	// 2. Redis 中查無資料，查 LeetCode GraphQL
	result, err := fetchLeetCodeQuestion(slug)
	if err == nil && result != nil {
		acRate := extractAcRate(result.Data.Question.Stats)
		estimated := estimateDifficulty(acRate)
		tags := extractTags(result.Data.Question.TopicTags)

		// 3. 寫入 Redis 快取
		rdb.Set(ctx, redisKey, fmt.Sprintf("%d", estimated), time.Hour)

		// 4. 封裝並送出任務到 RabbitMQ queue
		task := mq.QuestionTask{
			Slug:       slug,
			Title:      result.Data.Question.Title,
			Difficulty: result.Data.Question.Difficulty,
			Rating:     estimated,
			Tags:       tags,
		}
		mq.PublishQuestion(task)

		// 5. 同步回應給前端
		c.JSON(http.StatusOK, gin.H{
			"title":      task.Title,
			"difficulty": task.Difficulty,
			"rating":     task.Rating,
			"tags":       task.Tags,
		})
		return
	}

	// 6. 若連 LeetCode 都失敗，標示尚未分析
	log.Printf("[Queue] 題目 %s 尚未分析，排入佇列待處理", slug)
	_ = rdb.Set(ctx, redisKey, "", time.Minute*5).Err()
	c.JSON(http.StatusAccepted, gin.H{"status": "pending"})
}

// LeetCode 查詢用結構
type LeetCodeQuery struct {
	OperationName string                 `json:"operationName"`
	Variables     map[string]interface{} `json:"variables"`
	Query         string                 `json:"query"`
}

type LeetCodeResponse struct {
	Data struct {
		Question struct {
			QuestionID string `json:"questionId"`
			Title      string `json:"title"`
			Difficulty string `json:"difficulty"`
			TopicTags  []Tag  `json:"topicTags"` // ✅ 改為具名型別
			Stats      string `json:"stats"`
		} `json:"question"`
	} `json:"data"`
}

func fetchLeetCodeQuestion(slug string) (*LeetCodeResponse, error) {
	url := "https://leetcode.com/graphql"
	query := `query questionData($titleSlug: String!) {
		question(titleSlug: $titleSlug) {
			questionId
			title
			difficulty
			likes
			dislikes
			topicTags {
				name
				slug
			}
			stats
			content
		}
	}`

	payload := LeetCodeQuery{
		OperationName: "questionData",
		Variables:     map[string]interface{}{"titleSlug": slug},
		Query:         query,
	}

	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Referer", "https://leetcode.com/problems/"+slug+"/")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBytes, _ := io.ReadAll(resp.Body)
	var result LeetCodeResponse
	if err := json.Unmarshal(respBytes, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func extractAcRate(statsRaw string) float64 {
	fmt.Println("原始 statsRaw:", statsRaw)

	var intermediate string
	if err := json.Unmarshal([]byte(statsRaw), &intermediate); err == nil {
		statsRaw = intermediate
		fmt.Println("解包後 statsRaw:", statsRaw)
	}

	var stats map[string]interface{}
	if err := json.Unmarshal([]byte(statsRaw), &stats); err != nil {
		fmt.Println("解析 stats 錯誤:", err)
		return 0.0
	}

	acRateRaw, ok := stats["acRate"]
	if !ok {
		fmt.Println("找不到 acRate 欄位")
		return 0.0
	}

	// acRate 是 "55.7%" → 去掉 %
	acRateStr := fmt.Sprintf("%v", acRateRaw)
	acRateStr = strings.TrimSuffix(acRateStr, "%")

	acRate, err := strconv.ParseFloat(acRateStr, 64)
	if err != nil {
		fmt.Println("acRate 轉 float 失敗:", err)
		return 0.0
	}
	return acRate / 100.0
}

func estimateDifficulty(acRate float64) int {
	if acRate < 0.01 {
		acRate = 0.01
	}
	if acRate > 0.99 {
		acRate = 0.99
	}
	return int(1300 + 1700*(1-acRate))
}

func extractTags(tags []Tag) []string {
	out := make([]string, 0, len(tags))
	for _, t := range tags {
		if name := strings.TrimSpace(t.Name); name != "" {
			out = append(out, name)
		}
	}
	return out
}
