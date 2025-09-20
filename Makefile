APP_NAME := zed
SRC := cmd/main.go
OUTPUT := $(APP_NAME)

all: build

build: $(SRC)
	@echo "Building the project..."
	go build -o $(OUTPUT) $(SRC)

help:
	@echo "Makefile targets:"
	@echo "  build   			- Build the Go project"
	@echo "  help    			- Show this help message"