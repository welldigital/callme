run-mysql:
	docker start mysql-callme || docker run --rm --name mysql-callme -p 3306:3306 -e MYSQL_ROOT_PASSWORD=callme -d mysql:5.6

run-worker:
	cd ./worker/ && CALLME_CONNECTION_STRING='root:callme@tcp(localhost:3306)/callme?parseTime=true&multiStatements=true' go run main.go

run-api:
	cd ./api/ && go build && CALLME_CONNECTION_STRING='root:callme@tcp(localhost:3306)/callme?parseTime=true&multiStatements=true' ./api

run-integration-test:
	cd ./worker/ && go build && cd ../harness && go run main.go

docker-build:
	cd ./worker/ && GOOS=linux go build -o worker_linux
	docker build -t welldigital/callme:worker-latest ./worker/. 
	cd ./api/ && GOOS=linux go build -o api_linux
	docker build -t welldigital/callme:api-latest ./api/. 

docker-compose-up:
	docker-compose up --build

docker-compose-build: docker-build docker-compose-up

install:
	cd ./worker/ && dep ensure -v
	cd ./api/ && dep ensure -v
