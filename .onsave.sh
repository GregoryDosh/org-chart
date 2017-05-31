set -e

gofmt -s -w ./
go build . errors
go vet
golint
go build
./org-chart
dot -Tpng graph.dot -o Test\ PNG.png
dot -Tpdf graph.dot -o Test\ PDF.pdf
dot -Tsvg graph.dot -o Test\ SVG.svg
rm images/tmp*.jpg 2>/dev/null || true