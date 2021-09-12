APP := jsonnet-playground
ORG := jaredallard
_ := $(shell ./scripts/bootstrap-lib.sh) 

include .bootstrap/root/Makefile

###Block(targets)
.PHONY: watch
watch:
	@./scripts/shell-wrapper.sh gobin.sh github.com/cosmtrek/air@v1.27.3
###EndBlock(targets)
