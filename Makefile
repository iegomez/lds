all:
	go build

requirements:
	go get -u github.com/andlabs/ui/...
	go get -u github.com/BurntSushi/toml