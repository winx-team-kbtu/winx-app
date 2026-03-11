FROM golang:1.25-alpine AS base
WORKDIR /app

RUN apk add --no-cache ca-certificates git tzdata

COPY go.mod ./
RUN go mod download

COPY . ./

FROM base AS dev
RUN go install github.com/air-verse/air@latest
CMD ["air", "-c", "/app/air.toml"]

FROM base AS build
RUN CGO_ENABLED=0 go build -o main cmd/http/main.go

FROM alpine:latest AS publish
WORKDIR /app

RUN apk add --no-cache tzdata

ENV TZ=Asia/Almaty
RUN cp /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

COPY --from=build /app/main main
COPY --from=build /app/swagger swagger

ENTRYPOINT ["/app/main"]
