# ---------- BUILD ----------
FROM golang:1.25-alpine AS build
WORKDIR /app

RUN apk add --no-cache ca-certificates git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Only build main, no migrate
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main cmd/http/main.go

# ---------- RUNTIME ----------
FROM alpine:latest
WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata
ENV TZ=Asia/Almaty
RUN cp /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

COPY --from=build /app/main ./main
COPY --from=build /app/config.dev.development.yaml ./config.dev.development.yaml

EXPOSE 3004

CMD ["./main"]