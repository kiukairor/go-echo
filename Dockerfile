FROM --platform=linux/arm64 golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o go-echo .

FROM --platform=linux/arm64 alpine:3.21
WORKDIR /app
COPY --from=builder /app/go-echo .
EXPOSE 8080
CMD ["./go-echo"]
