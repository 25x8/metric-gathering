.PHONY: proto clean build test

# Генерация protobuf файлов
proto:
	protoc --go_out=internal/grpc/pb --go_opt=module=github.com/25x8/metric-gathering/internal/grpc/pb \
		--go-grpc_out=internal/grpc/pb --go-grpc_opt=module=github.com/25x8/metric-gathering/internal/grpc/pb \
		proto/metrics.proto

# Очистка сгенерированных файлов
clean-proto:
	rm -f internal/grpc/pb/*.pb.go

# Пересоздание protobuf файлов
regenerate-proto: clean-proto proto

# Сборка
build:
	go build -o bin/server ./cmd/server
	go build -o bin/agent ./cmd/agent

# Тесты
test:
	go test ./...

# Тесты с покрытием
test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Очистка всех артефактов
clean:
	rm -rf bin/
	rm -f coverage.out coverage.html
	rm -f *.pprof profiles/*.pprof

# Установка зависимостей для разработки
install-tools:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest 