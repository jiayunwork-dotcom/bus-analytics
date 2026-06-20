FROM golang:1.21-alpine

WORKDIR /app

RUN apk add --no-cache gcc musl-dev

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o bus-analytics ./cmd/server

EXPOSE 8080

CMD ["./bus-analytics"]
