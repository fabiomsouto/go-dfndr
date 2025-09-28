BUILD_FOLDER := ./bin
BIN := dfndr

.PHONY: run

run:
	go run -ldflags "-X main.version=`git describe --tags --always`" -o $(BUILD_FOLDER)/$(BIN) main.go
	$(BUILD_FOLDER)/$(BIN)