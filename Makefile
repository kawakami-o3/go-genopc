all: build

build:
	goimports -w -l .
	statik -src lib -m -f
	go build

test: build
	go test

clean:
	go clean
	rm -rf statik
	rm -f main_gen.go

deps:
	go get github.com/rakyll/statik

