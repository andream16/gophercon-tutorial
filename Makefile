.PHONY: tidy-vendor-all
tidy-vendor-all:
	@for dir in $$(find . -name "go.mod" -exec dirname {} \;); do \
		echo "Processing $$dir"; \
		cd "$$dir" && go mod tidy && go mod vendor && cd - > /dev/null; \
	done