.PHONY: gen-client-server

gen-client-server:
	go tool oapi-codegen \
		-o gopher/server/gen/gorilla.go \
		-generate=client,types,spec,gorilla \
		-package=openapi \
			../openapi.spec.yaml