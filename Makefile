.PHONY: tidy-vendor-all fmt

tidy-vendor-all:
	@echo "Preparing to vendor dependencies...";
	@for dir in $$(find . -name "go.mod" -exec dirname {} \;); do \
		echo "Processing $$dir"; \
		cd "$$dir" && go mod tidy && go mod vendor && cd - > /dev/null; \
	done
	@echo "Done vendoring dependencies!";

fmt:
	@echo "Preparing to format...";
	@for dir in $$(find . -name "go.mod" -exec dirname {} \;); do \
		echo "Processing $$dir"; \
		cd "$$dir" && go fmt ./... && cd - > /dev/null; \
	done
	@echo "Done formatting!";