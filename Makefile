
#
# Build WoofWoof
#

PROG=woofwoof

VERSION=$(shell cat ./VERSION)

# OS Detection
UNAME=$(shell uname)
ifeq ($(UNAME), Darwin)
  TARGET=osx
else
  TARGET=linux
endif

ARCH=amd64

LDFLAGS=-ldflags="-X main.version=$(VERSION)"

all: deps $(TARGET)
	@echo "Built $(VERSION) @ $(TARGET)"

deps:
	go get .

osx: 
	GOARCH=$(ARCH) GOOS=darwin go build $(LDFLAGS) -o $(PROG)

linux: 
	GOARCH=$(ARCH) GOOS=linux go build $(LDFLAGS) -o $(PROG)

