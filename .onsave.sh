set -e

gofmt -s -w ./
go build . errors
go vet
golint
go build
./org-chart
dot -Tpng graph.dot -o Test.png