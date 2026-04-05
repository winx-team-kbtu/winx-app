# ---------- BUILD STAGE ----------
FROM golang:1.25-alpine AS build

WORKDIR /app

RUN apk add --no-cache ca-certificates git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main cmd/http/main.go
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o migrate cmd/migration/main.go

# ---------- RUNTIME STAGE ----------
FROM alpine:latest

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

ENV TZ=Asia/Almaty
RUN cp /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

# Copy binary
COPY --from=build /app/main ./main
COPY --from=build /app/migrate ./migrate

# 🔥 CRITICAL: unified config name
COPY --from=build /app/config.dev.yaml ./config.dev.development.yaml

# (optional) swagger only for api-gateway
# COPY --from=build /app/swagger ./swagger

EXPOSE 3003

CMD ["sh", "-c", "./migrate -cmd up && ./main"]