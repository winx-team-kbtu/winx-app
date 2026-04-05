# ---------- BUILD STAGE ----------
FROM golang:1.25-alpine AS build

WORKDIR /app

RUN apk add --no-cache ca-certificates git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main cmd/http/main.go

FROM alpine:latest

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

ENV TZ=Asia/Almaty
RUN cp /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

COPY --from=build /app/main ./main
COPY --from=build /app/swagger ./swagger
COPY --from=build /app/config.dev.yaml ./config.dev.development.yaml

EXPOSE 3000

CMD ["./main"]