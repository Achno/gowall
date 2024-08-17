# Variables
BINARY_NAME = gowall
SOURCE_FILES = $(wildcard *.go)
MAN_DIR = /usr/local/share/man/man1
MAN_PAGE = $(BINARY_NAME).1

# Default target
all: build

# Build the Go binary
build: $(SOURCE_FILES)
	@echo "Building $(BINARY_NAME)..."
	@go build -o $(BINARY_NAME)

# Install the binary and the man page
install: build
	@echo "Installing $(BINARY_NAME) to /usr/local/bin..."
	@sudo install -m 0755 $(BINARY_NAME) /usr/local/bin/$(BINARY_NAME)
	@echo "Installing man page to $(MAN_DIR)..."
	@sudo install -m 0644 $(MAN_PAGE) $(MAN_DIR)/$(MAN_PAGE)
	@echo "Installation complete."

# Uninstall the binary and the man page
uninstall:
	@echo "Uninstalling $(BINARY_NAME) from /usr/local/bin..."
	@sudo rm -f /usr/local/bin/$(BINARY_NAME)
	@echo "Uninstalling man page from $(MAN_DIR)..."
	@sudo rm -f $(MAN_DIR)/$(MAN_PAGE)
	@echo "Uninstallation complete."

# Clean up build artifacts
clean:
	@echo "Cleaning up..."
	@rm -f $(BINARY_NAME)
	@echo "Cleanup complete."

# Generate the man page
man:
	@echo "Generating man page..."
	@man ./$(MAN_PAGE)

# Phony targets
.PHONY: all build install uninstall clean man
