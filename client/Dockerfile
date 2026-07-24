# Build Stage
FROM node:20-alpine AS builder
WORKDIR /app

# 패키지 파일 복사 및 설치
COPY package*.json ./
RUN npm install

# 소스코드 복사 및 Production 빌드
COPY . .
RUN npm run build

# Run Stage (Caddy를 사용한 Static 파일 서빙)
FROM caddy:2-alpine
COPY --from=builder /app/dist /srv

EXPOSE 80
CMD ["caddy", "file-server", "--root", "/srv", "--listen", ":80"]
