rootPath=$(cd "$(dirname "$0")";pwd)
GOOS=darwin GOARCH=amd64 go build -o ${rootPath}/bin/main ${rootPath}/main.go