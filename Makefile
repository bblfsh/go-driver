
test-native-internal:
	cd native; \
	go get -d -t ./... && \
	go test -v ./...

build-native-internal:
	cd native; \
	go get -d ./...  && \
	go build -o $(BUILD_PATH)/bin/native native.go

include .sdk/Makefile
