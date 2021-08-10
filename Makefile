.PHONY: build
PKGS := $(shell go list ./...)

install: build
	@go install xgc.go

fmt:
	go fmt $(PKGS)

build:
	@go mod vendor
