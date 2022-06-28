APP := jsonnet-playground
ORG := jaredallard
_ := $(shell ./scripts/bootstrap-lib.sh) 

#pre-build:: node-build gogenerate build

include .bootstrap/root/Makefile

###Block(targets)
.PHONY: watch
watch:
	@./scripts/shell-wrapper.sh gobin.sh github.com/cosmtrek/air@v1.40.2

.PHONY: node-build
node-build:
	@yarn build
###EndBlock(targets)
