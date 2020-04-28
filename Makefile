COMMIT_HOOK=.git/hooks/commit-msg

all: init
	go build -o gui


init: ${COMMIT_HOOK}
	
${COMMIT_HOOK}:
	(cd .git/hooks; ln -s ../../.githooks/commit-msg)

clean:
	go clean

tidy:
	go mod tidy -v

