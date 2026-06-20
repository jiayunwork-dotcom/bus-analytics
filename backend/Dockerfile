FROM golang:1.21-alpine

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .
RUN rm -f go.sum && go mod tidy

RUN CGO_ENABLED=0 go build -o /app/bus-analytics ./cmd/server

EXPOSE 8080

CMD ["/app/bus-analytics"]
