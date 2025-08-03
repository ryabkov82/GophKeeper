.PHONY: run-client run-server generate

run-client:
    go run ./cmd/client --tui  # Запуск TUI

run-server:
    go run ./cmd/server

generate:
    protoc --go_out=. --go-grpc_out=. proto/*.proto  # Генерация gRPC-кода