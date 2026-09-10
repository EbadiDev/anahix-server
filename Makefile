.PHONY: all build run test tidy clean swagger-tools swagger-gen swagger-clean

GOPATH ?= $(shell go env GOPATH)

all: build

build:
	@echo "🔨 Building Anahix Server..."
	go build -o bin/anahix-server main.go

run:
	@echo "🚀 Running Anahix Server..."
	go run main.go run

test:
	@echo "🧪 Running unit & integration tests..."
	go test -v ./...

tidy:
	@echo "📦 Tidying Go dependencies..."
	go mod tidy

clean:
	@rm -rf bin/

# Development Environment (Postgres + Redis)
dev-env-up:
	@echo "🐳 Starting PostgreSQL and Redis containers..."
	docker compose -f docker-compose.dev.yml up -d

dev-env-down:
	@echo "🛑 Stopping PostgreSQL and Redis containers..."
	docker compose -f docker-compose.dev.yml down

# Swagger/OpenAPI documentation (ArchNet pattern)
swagger-tools:
	@mkdir -p $(GOPATH)/bin
	@if [ ! -f $(GOPATH)/bin/goctl ]; then \
		echo "⬇️ Downloading goctl..."; \
		curl -C - --retry 5 --retry-delay 2 -fsSL https://github.com/zeromicro/go-zero/releases/download/tools%2Fgoctl%2Fv1.7.2/goctl-v1.7.2-linux-amd64.tar.gz -o /tmp/goctl.tar.gz && \
		tar -xzf /tmp/goctl.tar.gz -C $(GOPATH)/bin && \
		chmod +x $(GOPATH)/bin/goctl && \
		rm -f /tmp/goctl.tar.gz; \
	fi
	@if [ ! -f $(GOPATH)/bin/goctl-swagger ]; then \
		echo "⬇️ Downloading goctl-swagger..."; \
		curl -C - --retry 5 --retry-delay 2 -fsSL https://github.com/tensionc/goctl-swagger/releases/download/v1.0.1/goctl-swagger-v1.0.1-linux-amd64.tar.gz -o /tmp/goctl-swagger.tar.gz && \
		tar -xzf /tmp/goctl-swagger.tar.gz -C $(GOPATH)/bin && \
		chmod +x $(GOPATH)/bin/goctl-swagger && \
		rm -f /tmp/goctl-swagger.tar.gz; \
	fi

swagger-gen: swagger-tools
	@mkdir -p swagger
	@PATH="$(GOPATH)/bin:$$PATH" goctl api plugin -plugin goctl-swagger='swagger -filename anahix.json -pack Response -response "[{\"name\":\"code\",\"type\":\"integer\",\"description\":\"Status code\"},{\"name\":\"msg\",\"type\":\"string\",\"description\":\"Message\"},{\"name\":\"data\",\"type\":\"object\",\"description\":\"Data\",\"is_data\":true}]";' -api anahix.api -dir ./swagger
	@echo "✅ Swagger generated at swagger/anahix.json"

swagger-clean:
	@rm -rf swagger
