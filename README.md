# LeetCodeRaing
LeetCodeRating 是一套以 Golang 開發的評分與任務派發系統，支援 RabbitMQ 任務佇列、PostgreSQL 資料儲存、Redis 快取，並可透過 Kubernetes + GitLab CI/CD 自動部署。

## 📌 使用流程

1. 使用者透過前端頁面提交查詢請求
2. 先從 Redis 查詢 ， 若是存在則直接回傳結果
3. 若是不存在，則使用GraphQL及LeetCode API獲取該提相關Data，並將寫入任務送入 RabbitMQ queue
4. Worker 從 queue 消費任務，計算後寫入資料庫

## 專案架構
![image](https://github.com/user-attachments/assets/149a761a-daad-4324-bfb8-8505e68f0b3b)

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

本專案依賴以下外部服務，建議於本地或雲端環境事先配置：

| 服務       | 用途                        |
|------------|-----------------------------|
| Redis      | 快取存取、狀態管理            |
| PostgreSQL | 任務記錄與評分資料儲存        |
| RabbitMQ   | 任務排程與非同步訊息傳遞      |

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

- Go 1.23.9
- Gin Web Framework
- pgx (PostgreSQL driver)
- amqp (RabbitMQ)
- Redis (go-redis)
- Kubernetes / Minikube
- GitLab CI/CD
- Docker

---

## 📂 專案結構簡介

```text
.
├── api/                    # API 服務端點
├── worker/                 # 任務處理背景服務
├── k8s/                    # Kubernetes 部署檔案
├── Dockerfile              # API 映像建構檔
├── Dockerfile.worker       # Worker 映像建構檔
├── .gitlab-ci.yml          # GitLab CI/CD 腳本
└── README.md               # 本說明文件
