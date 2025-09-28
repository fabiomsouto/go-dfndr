BUILD_FOLDER := ./bin
BIN := dfndr

.PHONY: run

run:
	go build -o $(BUILD_FOLDER)/$(BIN) *.go
	$(BUILD_FOLDER)/$(BIN)