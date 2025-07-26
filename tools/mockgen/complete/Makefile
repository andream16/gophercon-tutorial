.PHONY: gen-mocks

gen-mocks:
	go tool mockgen -package gopher_test -source gopher/gopher.go -destination gopher/gopher_mock_test.go Storer \
		go mod tidy \
		go mod vendor