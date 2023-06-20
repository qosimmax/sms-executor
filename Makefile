CURRENT_DIR=$(shell pwd)

APP=$(shell basename ${CURRENT_DIR})

APP_CMD_DIR=${CURRENT_DIR}/cmd/server


build:
	CGO_ENABLED=0 GOOS=linux go build  -a -installsuffix cgo -o ${CURRENT_DIR}/bin/${APP} ${APP_CMD_DIR}/main.go

## server: runs the server
run:
	go run ${APP_CMD_DIR}/main.go

## test: runs tests
test:
	go test  ./...
## docker: Builds and runs the app via the project dockerfile, importing the .env-file as environment variables.
docker:
	docker build -t $(APP) .
	docker run --rm --name $(APP) -p 8000:8000 --env-file .env -it $(APP)
