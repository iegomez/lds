COMMIT_HOOK=.git/hooks/commit-msg
GOEXE ?= $(shell go env GOEXE)

all: init
	go build -o gui${GOEXE}

init: ${COMMIT_HOOK}
	
${COMMIT_HOOK}:
	(cd .git/hooks; ln -s ../../.githooks/commit-msg)

clean:
	go clean

tidy:
	go mod tidy -v

