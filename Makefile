build:
	@go build -o bin/go_bank_api

run: build
	@./bin/go_bank_api

test: 
	@go test -v ./...

clean:
	@rm ./bin/* ./tmp/*