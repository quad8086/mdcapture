GO=go
VERSION=dev
SRCDIR=$(shell pwd)
INSTALL_ROOT=$(SRCDIR)
INSTALL_DIR=$(INSTALL_ROOT)/.install
TARGET=mdcapture

all: deps build

deps:
	$(GO) mod download github.com/gorilla/websocket
	$(GO) mod download github.com/jessevdk/go-flags

build:
	$(GO) build

clean:
	$(RM) $(TARGET)

install:
	mkdir -p $(INSTALL_DIR)/bin
	install --mode 755 $(TARGET) $(INSTALL_DIR)/bin
