LOCAL_BIN := $(CURDIR)/bin

PROTOC := PATH="$$PATH:$(LOCAL_BIN)" protoc

PROTO_PATH := $(CURDIR)/api

PKG_PROTO_PATH := $(CURDIR)/pkg

VENDOR_PROTO_PATH := $(CURDIR)/vendor.protobuf

DB_DRIVER ?= postgres
DB_USER ?= postgres
DB_PASSWORD ?= postgres
DB_ADDRESS ?= localhost
DB_PORT ?= 5432
DB_NAME ?= logstream

.bin-deps: export GOBIN := $(LOCAL_BIN)
.bin-deps:
	$(info Installing binary dependencies...)

	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install github.com/pressly/goose/v3/cmd/goose@latest

.migrate-up: export GOOSE_DRIVER := $(DB_DRIVER)
.migrate-up: export GOOSE_DBSTRING := postgresql://$(DB_USER):$(DB_PASSWORD)@$(DB_ADDRESS):$(DB_PORT)/$(DB_NAME)?sslmode=disable
.migrate-up:
	$(info Migrating up...)

	goose -dir internal/database/migrations up

.migrate-down: export GOOSE_DRIVER := $(DB_DRIVER)
.migrate-down: export GOOSE_DBSTRING := postgresql://$(DB_USER):$(DB_PASSWORD)@$(DB_ADDRESS):$(DB_PORT)/$(DB_NAME)?sslmode=disable
.migrate-down:
	$(info Migrating down...)

	goose -dir internal/database/migrations down

.vendor-reset:
	rm -rf $(VENDOR_PROTO_PATH)
	mkdir -p $(VENDOR_PROTO_PATH)

.vendor-tidy:
	find $(VENDOR_PROTO_PATH) -type f ! -name "*.proto" -delete
	find $(VENDOR_PROTO_PATH) -empty -type d -delete

.vendor-google-protobuf:
	git clone -b main --single-branch -n --depth=1 --filter=tree:0 \
		https://github.com/protocolbuffers/protobuf $(VENDOR_PROTO_PATH)/protobuf/src/google/protobuf
	cd $(VENDOR_PROTO_PATH)/protobuf &&\
	git sparse-checkout set --no-cone src/google/protobuf &&\
	git checkout
	mkdir -p $(VENDOR_PROTO_PATH)/google
	mv $(VENDOR_PROTO_PATH)/protobuf/src/google/protobuf $(VENDOR_PROTO_PATH)/google/api
	rm -rf $(VENDOR_PROTO_PATH)/protobuf

.vendor-protovalidate:
	git clone -b main --single-branch --depth=1 --filter=tree:0 \
		https://github.com/bufbuild/protovalidate $(VENDOR_PROTO_PATH)/protovalidate/proto/protovalidate/buf
	cd $(VENDOR_PROTO_PATH)/protovalidate
	git checkout
	mv $(VENDOR_PROTO_PATH)/protovalidate/proto/protovalidate/buf $(VENDOR_PROTO_PATH)/buf
	rm -rf $(VENDOR_PROTO_PATH)/protovalidate

.vendor-googleapis:
	git clone -b main --single-branch -n --depth=1 --filter:tree:0 \
		https://github.com/googleapis/googleapis $(VENDOR_PROTO_PATH)/googleapis
	cd $(VENDOR_PROTO_PATH)/googleapis &&\
	git checkout
	mv $(VENDOR_PROTO_PATH)/googleapis/google $(VENDOR_PROTO_PATH)
	rm -rf $(VENDOR_PROTO_PATH)/googleapis

.protoc-generate:
	mkdir -p $(PKG_PROTO_PATH)
	$(PROTOC) --proto_path=$(CURDIR) \
	--go_out=$(PKG_PROTO_PATH) --go_opt paths=source_relative \
	--go-grpc_out=$(PKG_PROTO_PATH) --go-grpc_opt paths=source_relative \
	$(PROTO_PATH)/logstream/service.proto \
	$(PROTO_PATH)/logstream/messages.proto

.tidy:
	GOBIN=$(LOCAL_BIN) go mod tidy

.vendor: .vendor-reset .vendor-protovalidate .vendor-googleapis .vendor-tidy

# Генерация кода из protobuf
.generate: .bin-deps .protoc-generate .tidy

.build:
	go build -o $(LOCAL_BIN) ./cmd/logstream/client
	go build -o $(LOCAL_BIN) ./cmd/logstream/server