rootPath=$(cd "$(dirname "$0")";pwd)
echo $rootPath
./bin/monitoring -r "$rootPath" -c $rootPath/config.yaml