rootPath=$(cd "$(dirname "$0")";pwd)
go mod tidy
GOOS=linux GOARCH=amd64 go build -o ${rootPath}/bin/monitoring ${rootPath}/main.go
#GOOS=darwin GOARCH=amd64 go build -o ${rootPath}/bin/main ${rootPath}/main.go