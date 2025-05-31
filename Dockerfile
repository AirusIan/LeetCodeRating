# --- 建置階段 ---
FROM golang:1.23.9-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

# 複製並下載依賴
COPY go.mod go.sum ./
RUN go mod download

# 複製完整原始碼（包含子資料夾）
COPY . .

# 編譯成靜態執行檔1
RUN go build -o leetcoderating main.go

# --- 執行階段 ---
FROM alpine:latest AS api


WORKDIR /app
RUN apk add --no-cache ca-certificates

COPY --from=builder /app/leetcoderating .
COPY index.html . 

EXPOSE 8080

CMD ["./leetcoderating"]
