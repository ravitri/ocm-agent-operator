FIPS_ENABLED=true
include boilerplate/generated-includes.mk

OPERATOR_NAME=ocm-agent-operator
CLI_NAME=ocm-agent

.PHONY: boilerplate-update
boilerplate-update: ## Make boilerplate update itself
	@boilerplate/update

# Run against the configured Kubernetes cluster in ~/.kube/config
.PHONY: run
run:
	OPERATOR_NAMESPACE="openshift-ocm-agent-operator" go run ./main.go

.PHONY: run-verbose
run-verbose:
	OPERATOR_NAMESPACE="openshift-ocm-agent-operator" go run ./main.go --zap-log-level=5

.PHONY: tools
tools: ## Install local go tools for OAO
	cat tools.go | grep _ | awk -F'"' '{print $$2}' | xargs -tI % go install %

.PHONY: ocm-agent-build
ocm-agent-build: ## Build ocm-agent binary
	GOOS=linux go build -o build/ocm-agent/_output/bin/$(CLI_NAME) ./cmd/ocm-agent

.PHONY: clean
clean:
	rm -f build/_output/bin/* build/ocm-agent/_output/bin/*

.PHONY: help
help: ## Show this help screen.
		@echo 'Usage: make <OPTIONS> ... <TARGETS>'
		@echo ''
		@echo 'Available targets are:'
		@echo ''
		@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | sed 's/##//g' | awk 'BEGIN {FS = ":"}; {printf "\033[36m%-30s\033[0m %s\n", $$2, $$3}'
