FROM golang:1.22-alpine AS build
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o tracker ./cmd/tracker

FROM alpine:3.19
COPY --from=build /app/tracker /usr/local/bin/tracker
COPY config.yml /etc/tracker/config.yml

EXPOSE 8080
ENTRYPOINT ["tracker", "-config", "/etc/tracker/config.yml"]
