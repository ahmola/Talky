# Build Stage
FROM golang:1.22-alpine AS builder
WORKDIR /app

# 의존성 캐싱
COPY go.mod go.sum ./
RUN go mod download

# 소스코드 및 공통 모듈 복사
COPY proto/ proto/
COPY chat-service/ chat-service/

# 빌드 실행
RUN CGO_ENABLED=0 GOOS=linux go build -o chat-service-bin ./chat-service/main.go

# Run Stage
FROM alpine:3.19
WORKDIR /app

# 빌드 결과물 복사
COPY --from=builder /app/chat-service-bin .

EXPOSE 8081
CMD ["./chat-service-bin"]
