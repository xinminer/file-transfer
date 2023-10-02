# Target name (executable file)
TARGET = client

# Paths and variables
SRC_DIR = ../cmd/client
BUILD_DIR = ./
GO = go

# Cleaning generated files
.PHONY: clean
clean:
	@echo "Cleaning the executable file..."
	@rm -f $(TARGET)

# Building the executable file
.PHONY: build
build: clean
	@echo "Building the executable file..."
	$(GO) mod download
	$(GO) build $(BUILD_FLAGS) -o $(TARGET) $(SRC_DIR)

# Running the built executable file
.PHONY: run
run: build
	@echo "Running the executable file..."
	./$(TARGET) -address $(ADDRESS) -port $(PORT)

# Help (display available commands)
.PHONY: help
help:
	@echo "Available commands:"
	@echo "    make clean"
	@echo "        Clean generated files"
	@echo "    make build"
	@echo "        Build the executable file"
	@echo "    make run ADDRESS=<string> PORT=<int>"
	@echo "        Run the built executable file"
	@echo "    make help"
	@echo "        Display this message"

# Default action is to display help
.DEFAULT_GOAL := help
