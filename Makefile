run-mysql:
	docker start mysql-callme || docker run --rm --name mysql-callme -p 3309:3306 -e MYSQL_ROOT_PASSWORD=callme -d mysql:5.6

run-cmd:
	# Server=localhost;Port=3309;Database=callme;Uid=root;Pwd=callme;AllowUserVariables=true;multiStatements=true;Charset=utf8
	cd ./cmd/ && CALLME_CONNECTION_STRING='root:callme@tcp(localhost:3309)/callme?parseTime=true&multiStatements=true' go run main.go

run-integration-test:
	cd ./cmd/ && go build && cd ../harness && go run main.go
