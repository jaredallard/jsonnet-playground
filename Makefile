APP := jsonnet-playground
ORG := rgst-io
_ := $(shell ./scripts/bootstrap-lib.sh) 

include .bootstrap/root/Makefile

.PHONY: watch
watch:
	@./scripts/shell-wrapper.sh gobin.sh github.com/cosmtrek/air@v1.40.2
