BINARY_NAME=home-ddns

build:
	GOARCH=amd64 GOOS=linux go build -o ${BINARY_NAME} main.go

run:
	./${BINARY_NAME}

build_and_run:
	build run

clean:
	go clean
	rm ${BINARY_NAME}
