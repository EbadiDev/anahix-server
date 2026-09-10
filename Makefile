.PHONY: all build run test tidy clean

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
