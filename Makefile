default: test

ci: test race

test:
	go test ./... -coverprofile=coverage.out -covermode=count -count=1

race:
	go test ./... -race -count=1 -run Test

lint:
	golangci-lint run ./...

depsdev:
	go install github.com/Songmu/ghch/cmd/ghch@latest
	go install github.com/Songmu/gocredits/cmd/gocredits@latest

prerelease:
	git pull origin main --tag
	go mod tidy
	ghch -w -N ${VER}
	gocredits . w
	git add CHANGELOG.md CREDITS go.mod go.sum
	git commit -m'Bump up version number'
	git tag ${VER}

prerelease_for_tagpr:
	gocredits . -w
	git add CHANGELOG.md CREDITS go.mod go.sum

.PHONY: default test
