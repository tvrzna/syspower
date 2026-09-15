DISTFILE=syspower
BUILD_VERSION=`git describe --tags`

clean:
	rm -r dist

build: build-cli build-gui

build-cli:
	mkdir -p dist
	go build -o dist/${DISTFILE} -ldflags "-X github.com/tvrzna/syspower/internal/syspower.buildVersion=${BUILD_VERSION}" -buildvcs=false ./cmd/syspower-cli

build-gui:
	mkdir -p dist
	go build -o dist/${DISTFILE}-gui -ldflags "-X github.com/tvrzna/syspower/internal/syspower.buildVersion=${BUILD_VERSION}" -buildvcs=false ./cmd/syspower-gui

install:
	install -DZs dist/${DISTFILE} ${DESTDIR}/usr/bin