all:
	go build

requirements:
	dep ensure -v

dev-requirements:
	go get -u github.com/golang/dep/cmd/dep