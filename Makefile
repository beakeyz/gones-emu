
BUILD_DIR = build
OUT = $(BUILD_DIR)/gones

build:
	@go build -C cmd -o ../$(OUT)

debug: build
	@./$(OUT)

clean:
	@rm $(BUILD_DIR)

.PHONY: build debug clean