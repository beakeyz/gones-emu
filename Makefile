
BUILD_DIR = build
VENDOR_DIR = ./vendor
OUT = $(BUILD_DIR)/gones

build:
	go build -C cmd -o ../$(OUT)

debug: build
	@./$(OUT)

clean:
	@rm -r $(BUILD_DIR)
	@rm -r $(VENDOR_DIR)
	@go mod vendor

.PHONY: build debug clean
