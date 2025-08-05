.PHONY: run-client run-server generate

run-client:
	go run ./cmd/client --tui  # Запуск TUI

run-server:
	go run ./cmd/server

generate:
	protoc --go_out=. \
	--go-grpc_out=. \
	--go-grpc_opt=module=github.com/ryabkov82/gophkeeper \
	--go_opt=module=github.com/ryabkov82/gophkeeper \
	--proto_path=internal/pkg/proto \
	internal/pkg/proto/api.proto