set -e

gofmt -s -w ./
go build . errors
go vet
golint
go build
./connected-graph
dot -Tpng graph.dot -o Test.png