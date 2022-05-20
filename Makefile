BINARY_NAME=home-ddns
DOCKER_IMAGE=home-ddns

build:
	GOARCH=amd64 GOOS=linux go build -o ${BINARY_NAME} main.go

run:
	./${BINARY_NAME}

build_and_run:
	build run

docker-build:
	docker build -t ${DOCKER_IMAGE} .

# Useful for testing
docker-run:
	docker run -v ${PWD}/config.yaml:/home-ddns/config.yaml -it home-ddns -v -cron -interval 1

clean:
	go clean
	rm ${BINARY_NAME}
