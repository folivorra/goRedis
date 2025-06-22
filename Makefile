BUF_VERSION                 = v1.54.0
PROTOC_GEN_GO_VERSION       = v1.36.6
PROTOC_GEN_GO_GRPC_VERSION  = v1.5.1
BIN_DIR          			= bin
BUF              			= $(BIN_DIR)/buf
PROTOC_GEN_GO    			= $(BIN_DIR)/protoc-gen-go
PROTOC_GEN_GO_GRPC 			= $(BIN_DIR)/protoc-gen-go-grpc

.PHONY: all gen clean

all: $(BUF) $(PROTOC_GEN_GO) $(PROTOC_GEN_GO_GRPC)

# ---------- buf --------------------------------------------------------------
$(BUF):
	@mkdir -p $(BIN_DIR)
	@echo "⏬  downloading buf $(BUF_VERSION)"
	@GOBIN=$(BIN_DIR) go install github.com/bufbuild/buf/cmd/buf@$(BUF_VERSION)
	@echo "✓ buf installed → $(BUF)"

# ---------- protoc-плагины ----------------------------------------------------
$(PROTOC_GEN_GO):
	@mkdir -p $(BIN_DIR)
	@echo "⏬  installing protoc-gen-go $(PROTOC_GEN_GO_VERSION)"
	@GOBIN=$(BIN_DIR) go install \
	    google.golang.org/protobuf/cmd/protoc-gen-go@$(PROTOC_GEN_GO_VERSION)

$(PROTOC_GEN_GO_GRPC):
	@mkdir -p $(BIN_DIR)
	@echo "⏬  installing protoc-gen-go-grpc $(PROTOC_GEN_GO_GRPC_VERSION)"
	@GOBIN=$(BIN_DIR) go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@$(PROTOC_GEN_GO_GRPC_VERSION)

# ---------- генерация через buf ----------------------------------------------
gen: $(BUF) $(PROTOC_GEN_GO) $(PROTOC_GEN_GO_GRPC)
	@mkdir -p $(GEN_DIR)
	@echo "⚙️  buf generate → $(GEN_DIR)"
	@$(BUF) generate
	@echo "✓ protobuf code generated"

# ---------- очистка -----------------------------------------------------------
clean:
	@rm -rf $(BIN_DIR) $(GEN_DIR)
	@echo "🧹  bin/ и proto/gen очищены"
