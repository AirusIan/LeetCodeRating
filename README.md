# LeetCodeRaing
LeetCodeRating 是一套用 Golang 開發的評分與任務派發系統，整合 RabbitMQ、PostgreSQL、Redis Cluster，搭配 Kubernetes + GitLab CI/CD 自動部署，支援Redis讀寫分離、壓力測試與 HPA 水平擴展。

## 📌 使用流程

1. 使用者透過前端頁面提交查詢請求
2. 先從 Redis Cluster 查詢快取結果
3. 若是不存在，則使用GraphQL及LeetCode API獲取該提相關Data，並將寫入任務送入 RabbitMQ queue
4. Worker 從 queue 消費任務，計算後寫入資料庫

## 專案架構
![image](https://github.com/user-attachments/assets/9a0f3234-d2f1-407f-a0ff-c2260f0cfdbf)


本系統包含下列核心組件：

### 🔹 API Service (`api`)
- 提供外部 HTTP 介面
- 接收任務、查詢進度、管理評分等操作
- 與 Redis、RabbitMQ、PostgreSQL 整合

### 🔹 Worker Service (`worker`)
- 接收來自 RabbitMQ 的任務訊息
- 執行評分邏輯、更新資料庫狀態
- 與 Redis 快取協作提升效率

---

## ⚙️ 相依服務

本專案依賴以下外部服務，建議於本地 Kubernetes 或雲端環境中事先部署：

| 服務           | 用途                              |
|----------------|-----------------------------------|
| Redis Cluster  | 查詢快取、減少 API 讀取壓力        |
| PostgreSQL     | 題目評分資料與查詢結果儲存        |
| RabbitMQ       | 任務佇列管理、解耦 API 與 Worker   |
| Ingress NGINX  | 統一入口，處理 URL rewrite 與流量控制 |


## ⚙️ CI/CD 自動化部署流程

本專案採用 **GitLab CI/CD** 搭配 **本機 Minikube** 進行容器化部署，模擬真實部署流程並熟悉 DevOps 工具鏈整合。

### CI/CD架構圖
![image](https://github.com/user-attachments/assets/d18f03d8-042e-4151-8321-6bca8dbb2288)


### ✅ 流程概述

1. 開發者推送程式碼到 GitLab
2. GitLab Runner（Shell Executor）在本機觸發 `.gitlab-ci.yml` 任務
3. 建置 API / Worker 映像並推送至 GitLab Container Registry
4. Minikube Pod 啟動並透過 `Secret` 取得 Registry 權限
5. 使用 `kubectl` 自動部署至本機 Minikube

### 📦 映像建置與推送

推送至：
`registry.gitlab.com/seang38077-group/leetcoderating/`

- leetcoderating-api:latest
- leetcoderating-worker:latest

## 🧱 使用技術

Go 1.23.9 + Gin

Redis Cluster (go-redis/v9)

RabbitMQ (amqp)

PostgreSQL (pgx)

Kubernetes + HPA + Ingress NGINX

GitLab CI/CD

k6 (Spike / Load test 工具)

---

## 📂 專案結構簡介

```text
.
.
├── api/
├── worker/
├── k6/
├── k8s/
├── Dockerfile
├── Dockerfile.worker
├── .gitlab-ci.yml
└── README.md

---

```
## 📈 Spike Test 改善過程與結果

專案初期進行 Spike Test（瞬間高併發測試）時，發現：

- API Service fail rate 達 40% 以上
- 部分 request latency 超過 3 秒
- Ingress NGINX 無流量限制，流量直接打爆後端

### 🛠️ 調整與改善步驟

1. Redis 由單節點改為 Redis Cluster
2. 調整 Redis Client 連線池（PoolSize、MinIdleConns、Timeout）
3. API Service 寫入 Redis 改為 Goroutine async 寫入
4. Ingress NGINX 增加：
    - `limit-connections`
    - `limit-rps`
5. API Service 加入 HPA（Horizontal Pod Autoscaler）
6. Spike Test 改進腳本：多樣化 slug 測試，模擬實際流量

---
## 🗺️ Redis Cluster 架構說明

目前 Redis 採用 Cluster 架構，每個 API Pod 透過 Redis Client 自動分配請求至不同 Redis Master 節點，實現讀寫流量分散，提升系統高併發承載能力。

![image](https://github.com/user-attachments/assets/e0bd864c-6a87-4c0b-9d29-b943027a5be6)


- Ingress 將流量平均分配至多個 API Pod
- 每個 Pod 透過 go-redis/v9 Cluster Client 自動管理連線池
- Redis Cluster 採 Master-Slave 結構，支援故障容錯
- Key slot 分散於多組 Master，避免單點瓶頸

✅ Spike Test 中觀察到 Redis Cluster 可有效分散高峰流量，降低 API timeout 機率，提升系統穩定性。


### 🚀 Spike Test 結果前後比較

| 項目                | 調整前             | 調整後             |
|---------------------|--------------------|--------------------|
| Spike Test fail rate | 約 40%             | 約 8%             |
| 平均 latency        | 約 900 ms ~ 1.5 s  | 約 700 ms ~ 1.2 s  |
| 高峰 latency        | > 3 秒             | < 1.8 秒           |
| Redis Load          | 單點，瓶頸明顯      | Cluster，平均分散  |
| API Pod             | 無 HPA，固定 1 pod  | HPA，2~5 自動擴展  |

---

目前系統可穩定處理 500 VUs Spike Test，fail rate 明顯降低，整體服務可用性提升。  
已支援更高並發及短時間高峰流量，後續可進一步優化 Redis Cluster Rebalance 及 Worker 擴展策略。
