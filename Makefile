
PACKAGE=github.com/troykinsella/bacon
BINARY=bacon
VERSION=0.0.6

LDFLAGS=-ldflags "-X main.AppVersion=${VERSION}"

build:
	go build ${LDFLAGS} ${PACKAGE}

install:
	go install ${LDFLAGS}

test:
	go test ${PACKAGE}/...

dist:
	gox ${LDFLAGS} \
		-arch="amd64" \
		-os="darwin linux windows" \
		-output="${BINARY}_{{.OS}}_{{.Arch}}" \
		${PACKAGE}

clean:
	test -f ${BINARY} && rm ${BINARY} || true
	rm ${BINARY}_* || true