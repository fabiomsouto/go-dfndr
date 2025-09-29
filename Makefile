BUILD_FOLDER := ./bin
BIN := dfndr

.PHONY: tidy

clean:
	rm -rf $(BUILD_FOLDER)

tidy: clean
	go mod tidy

build: tidy
	mkdir -p $(BUILD_FOLDER)
	go build -o $(BUILD_FOLDER)/$(BIN) *.go

run: build
	$(BUILD_FOLDER)/$(BIN)