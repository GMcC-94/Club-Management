# Build stage
FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git

# Install templ
RUN go install github.com/a-h/templ/cmd/templ@latest
ENV PATH="/go/bin:${PATH}"

WORKDIR /app

# Copy only go.mod - tidy will generate go.sum
COPY go.mod ./
RUN go mod download || true

# Copy source code
COPY . .

# Generate templ files first so go mod tidy can see all imports
RUN /go/bin/templ generate
RUN go mod tidy

# Build binaries
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/server
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o setup ./cmd/setup

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

COPY --from=builder /app/main .
COPY --from=builder /app/setup .

EXPOSE 8080

CMD ["./main"]
