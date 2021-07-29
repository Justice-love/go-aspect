PKGS := $(shell go list ./...)
install:
	@go mod vendor
	@go install xgc.go

fmt:
	go fmt $(PKGS)