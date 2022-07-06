rootPath=$(cd "$(dirname "$0")";pwd)
echo $rootPath
./bin/main -r "$rootPath" -c $rootPath/config.yaml