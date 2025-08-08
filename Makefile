.PHONY: run-client run-server generate test coverage-html

run-client:
	go run ./cmd/client --tls=true --skip-verify=false --log-level='debug'

run-server:
	go run ./cmd/server --s=true

generate:
	protoc --go_out=. \
	--go-grpc_out=. \
	--go-grpc_opt=module=github.com/ryabkov82/gophkeeper \
	--go_opt=module=github.com/ryabkov82/gophkeeper \
	--proto_path=internal/pkg/proto \
	internal/pkg/proto/api.proto


test:
	@echo "Running tests on packages:"
ifeq ($(OS),Windows_NT)
	@powershell -NoProfile -Command \
	"$$pkgs = (go list ./... | Select-String -NotMatch 'internal/pkg/proto' | ForEach-Object { $$_.Line }); \
	Write-Host 'Packages:' $$pkgs; \
	go test $$pkgs -coverprofile='coverage.out'"

else
	@pkgs=$$(go list ./... | grep -v internal/pkg/proto); \
	echo "Packages: $$pkgs"; \
	go test $$pkgs -coverprofile=coverage.out
endif
	go tool cover -func=coverage.out


coverage-html: test
	go tool cover -html=coverage.out -o coverage.html
