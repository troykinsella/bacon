BINARY=bacon
VERSION=1.0.0

LDFLAGS=-ldflags "-X main.AppVersion=${VERSION}"

setup:
	go install github.com/mitchellh/gox@latest

build:
	go mod vendor
	go build ${LDFLAGS} .

test:
	go test -cover ./...

dist:
	gox ${LDFLAGS} \
		-arch="amd64" \
		-os="darwin linux windows" \
		-output="${BINARY}_{{.OS}}_{{.Arch}}" \
		.
	sha256sum bacon_* > sha256sum.txt

clean:
	rm ${BINARY} || true
	rm ${BINARY}_* || true
	rm sha256sum.txt || true

.PHONY: setup build test dist clean
