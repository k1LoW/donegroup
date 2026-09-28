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

credits:
	go install github.com/Songmu/gocredits/cmd/gocredits@v1.0.0
	gocredits . > CREDITS

prerelease:
	git pull origin main --tag
	go mod tidy
	ghch -w -N ${VER}
	$(MAKE) credits
	git add CHANGELOG.md CREDITS go.mod go.sum
	git commit -m'Bump up version number'
	git tag ${VER}

prerelease_for_tagpr:
	$(MAKE) credits
	git add CHANGELOG.md CREDITS go.mod go.sum

.PHONY: default test credits
