FROM golang:1.14 as builder

WORKDIR /app

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o subscriber

FROM alpine:3.11
WORKDIR /app
COPY --from=builder /app/subscriber .

ENTRYPOINT ["/app/subscriber"]
