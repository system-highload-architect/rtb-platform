# --- CLUSTER BUF AUTOMATION ENGINE (OPERATOR CLASS) ---

.PHONY: all lint generate breaking sync clean

all: lint generate sync

# 1. Линтинг контрактов
lint:
	@echo ">> [BUF LINT]: Checking protobuf contracts syntax and style guidelines..."
	buf lint

# 2. Наносекундная генерация Go gRPC-мостов без protoc-грязи
generate:
	@echo ">> [BUF GENERATE]: Executing blazing-fast binary code generation..."
	@mkdir -p pb
	buf generate

# 3. Защита API от ломающих изменений (Backward Compatibility)
breaking:
	@echo ">> [BUF BREAKING]: Verifying contracts against breaking changes..."
	buf breaking --against '.git#branch=main'

# 4. Жесткая синхронизация именованных графов импортов Go Workspaces
sync:
	@echo ">> [GO WORKSPACES]: Synchronizing mono-repository dependency graphs..."
	@go work sync
	@echo ">> [SUCCESS]: Payment cluster binary stubs successfully scaffolded."

clean:
	@echo ">> [CLEAN]: Wiping out obsolete generation artifacts..."
	rm -rf pb/*
