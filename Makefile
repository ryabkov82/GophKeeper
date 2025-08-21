SHELL := bash
.SHELLFLAGS := -lc

.PHONY: run-client run-server generate test coverage-html mocks

run-client:
	go run ./cmd/client --tls=true --skip-verify=false --log-level='debug'

run-server:
	go run ./cmd/server --s=true --binary-path='E:\gophkeeper\binary_data'

generate:
	protoc --go_out=. \
	--go-grpc_out=. \
	--go-grpc_opt=module=github.com/ryabkov82/gophkeeper \
	--go_opt=module=github.com/ryabkov82/gophkeeper \
	--proto_path=internal/pkg/proto \
	internal/pkg/proto/api.proto


test:
	@echo "Running tests on packages:"
#ifeq ($(OS),Windows_NT)
#	@powershell -NoProfile -Command \
#	"$$pkgs = (go list ./... | Select-String -NotMatch 'internal/pkg/proto' | ForEach-Object { $$_.Line }); \
#	Write-Host 'Packages:' $$pkgs; \
#	go test $$pkgs -coverprofile='coverage.out'"
#
#else
	@pkgs=$$(go list ./... | grep -v internal/pkg/proto); \
	echo "Packages: $$pkgs"; \
	go test $$pkgs -coverprofile=coverage.out
#endif
	go tool cover -func=coverage.out


coverage-html: test
	go tool cover -html=coverage.out -o coverage.html

MOCKGEN=mockgen
PROTO_DIR=internal/pkg/proto
MOCKS_DIR=internal/pkg/proto/mocks

mocks:
	$(MOCKGEN) -source=$(PROTO_DIR)/api_grpc.pb.go -destination=$(MOCKS_DIR)/mock_api.go -package=mocks


##### ==== Cross‑platform builds with version metadata ==== #####

# ----- Настройки путей (проверьте!) -----
CLIENT_MAIN ?= ./cmd/client
SERVER_MAIN ?= ./cmd/server

# Полный импорт‑путь до пакета с клиентскими переменными (buildVersion/buildDate/buildCommit)
TUI_PKG ?= github.com/ryabkov82/gophkeeper/client/tui

# Куда вкалывать серверные переменные. По умолчанию — в пакет main.
SERVER_PKG ?= main

BIN_DIR ?= bin
GO ?= go

# ----- Матрица платформ -----
GOOS_LIST  ?= linux darwin windows
GOARCH_LIST ?= amd64 arm64

# ----- Версионные метаданные -----
# Если HEAD помечен тегом — используем тег,
# иначе — человеческое описание (ближайший тег + кол-во коммитов + sha),
# а если тегов совсем нет — короткий sha.
VERSION ?= $(shell \
  (git describe --tags --exact-match 2>/dev/null) || \
  (git describe --tags --dirty --always 2>/dev/null) || \
  (git rev-parse --short HEAD 2>/dev/null) )
ifeq ($(strip $(VERSION)),)
VERSION := $(shell git rev-parse --short HEAD)
endif
DATE   ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
COMMIT ?= $(shell git rev-parse --short HEAD)

# ----- LDFLAGS -----
LDFLAGS_COMMON := -s -w
LDFLAGS_CLIENT := $(LDFLAGS_COMMON) \
	-X '$(TUI_PKG).buildVersion=$(VERSION)' \
	-X '$(TUI_PKG).buildDate=$(DATE)' \
	-X '$(TUI_PKG).buildCommit=$(COMMIT)'

LDFLAGS_SERVER := $(LDFLAGS_COMMON) \
	-X '$(SERVER_PKG).buildVersion=$(VERSION)' \
	-X '$(SERVER_PKG).buildDate=$(DATE)' \
	-X '$(SERVER_PKG).buildCommit=$(COMMIT)'

# ----- Утилиты/хелперы -----
# rm, совместимый с bash (Git Bash/WSL). Для PowerShell используйте 'go clean' или удаление вручную.
RM ?= rm -rf

# Формирование имени файла с .exe для Windows
define BIN_NAME
$(BIN_DIR)/$(1)_$(2)_$(3)$(if $(filter $(1),windows),.exe,)
endef

# Унифицированная сборка одной цели
define BUILD_ONE
	@mkdir -p $(BIN_DIR)
	@echo "==> GOOS=$(1) GOARCH=$(2) $($(3)_MAIN) -> $(call BIN_NAME,$(1),$(2),$(4))"
	GOOS=$(1) GOARCH=$(2) CGO_ENABLED=0 \
		$(GO) build -ldflags "$($(5))" -o $(call BIN_NAME,$(1),$(2),$(4)) $($(3)_MAIN)
endef

.PHONY: client client-all client-linux client-darwin client-windows \
        server server-all server-linux server-darwin server-windows \
        clean print-version

# ===== Клиент =====
client-all:
	@for goos in $(GOOS_LIST); do \
	  for goarch in $(GOARCH_LIST); do \
	    $(MAKE) --no-print-directory _client_one GOOS=$$goos GOARCH=$$goarch; \
	  done; \
	done

_client_one:
	$(call BUILD_ONE,$(GOOS),$(GOARCH),CLIENT,client,LDFLAGS_CLIENT)

client-linux:
	$(MAKE) client-all GOOS_LIST="linux"
client-darwin:
	$(MAKE) client-all GOOS_LIST="darwin"
client-windows:
	$(MAKE) client-all GOOS_LIST="windows"

# ===== Сервер =====
server-all:
	@for goos in $(GOOS_LIST); do \
	  for goarch in $(GOARCH_LIST); do \
	    $(MAKE) --no-print-directory _server_one GOOS=$$goos GOARCH=$$goarch; \
	  done; \
	done

_server_one:
	$(call BUILD_ONE,$(GOOS),$(GOARCH),SERVER,server,LDFLAGS_SERVER)

server-linux:
	$(MAKE) server-all GOOS_LIST="linux"
server-darwin:
	$(MAKE) server-all GOOS_LIST="darwin"
server-windows:
	$(MAKE) server-all GOOS_LIST="windows"

# ----- Вспомогательные цели -----
print-version:
	@echo "VERSION=$(VERSION)"
	@echo "DATE=$(DATE)"
	@echo "COMMIT=$(COMMIT)"
	@echo "TUI_PKG=$(TUI_PKG)"
	@echo "SERVER_PKG=$(SERVER_PKG)"

clean:
	$(RM) $(BIN_DIR)
