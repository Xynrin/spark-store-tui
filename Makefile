VERSION := 0.8.0
PKGNAME := spark-store-tui
PKGROOT := package-root
ARCH ?= $(shell dpkg --print-architecture 2>/dev/null || echo amd64)
DEB := $(PKGNAME)_$(VERSION)-1_$(ARCH).deb
SOURCE_TAR := $(PKGNAME)-source-$(VERSION).tar.gz
SOURCE_DIR := $(PKGNAME)-source-$(VERSION)

.PHONY: all check build source clean local-install go-test go-build

all: check build

check:
	sh -n $(PKGROOT)/usr/bin/sparkstore
	sh -n $(PKGROOT)/DEBIAN/postinst
	sh -n $(PKGROOT)/DEBIAN/postrm
	go test ./...
	go vet ./...

# Build a native Debian package. The package stage is copied into Linux tmpfs
# first so this works when the checkout lives under WSL's /mnt/c (where Unix
# ownership and execute bits cannot be represented reliably).
build:
	@stage=$$(mktemp -d); \
	trap 'rm -rf "$$stage"' EXIT; \
	cp -a $(PKGROOT) "$$stage/$(PKGROOT)"; \
	find "$$stage/$(PKGROOT)" -type d -exec chmod 0755 {} +; \
	find "$$stage/$(PKGROOT)" -type f -exec chmod 0644 {} +; \
	rm -f "$$stage/$(PKGROOT)/usr/bin/spark-store-tui"; \
	mkdir -p "$$stage/$(PKGROOT)/usr/lib/sparkstore"; \
	goarch=$$(case "$(ARCH)" in amd64) echo amd64 ;; arm64) echo arm64 ;; *) echo "$(ARCH)" ;; esac); \
	GOOS=linux GOARCH=$$goarch CGO_ENABLED=0 go build -buildvcs=false -o "$$stage/$(PKGROOT)/usr/lib/sparkstore/sparkstore" ./cmd/spark-store-tui; \
	find "$$stage/$(PKGROOT)" -type d -exec chmod 0755 {} +; \
	chmod 0755 "$$stage/$(PKGROOT)/usr/bin/sparkstore" "$$stage/$(PKGROOT)/usr/lib/sparkstore/sparkstore" "$$stage/$(PKGROOT)/DEBIAN/postinst" "$$stage/$(PKGROOT)/DEBIAN/postrm"; \
	sed -i "s/^Version: .*/Version: $(VERSION)-1/; s/^Architecture: .*/Architecture: $(ARCH)/" "$$stage/$(PKGROOT)/DEBIAN/control"; \
	(cd "$$stage/$(PKGROOT)" && find usr -type f -print0 | sort -z | xargs -0 md5sum > DEBIAN/md5sums); \
	dpkg-deb --build --root-owner-group "$$stage/$(PKGROOT)" $(DEB)

source:
	@temporary=$$(mktemp); \
	tar --exclude='.git' --exclude='*.deb' --exclude='*.rpm' --exclude='*.tar.gz' --exclude='.cache' --exclude='build' --exclude='dist' --exclude='apt' --exclude='rpm' --exclude='package-root/tmp' --exclude='./packaging/aur/PKGBUILD' --exclude='./packaging/aur/.SRCINFO' --transform='s,^,$(SOURCE_DIR)/,' -czf "$$temporary" .; \
	mv "$$temporary" $(SOURCE_TAR)

local-install: build
	sudo apt install ./$(DEB)

go-test:
	go test ./...
	go vet ./...

go-build:
	mkdir -p build
	go build -buildvcs=false -o build/sparkstore ./cmd/spark-store-tui
	@probe=$$(mktemp -d build/.sparkstore-case.XXXXXX); \
	touch "$$probe/sparkstore"; \
	if [ ! -e "$$probe/SparkStore" ]; then \
		ln -sf sparkstore build/SparkStore; \
		ln -sf sparkstore build/SPARKSTORE; \
		ln -sf sparkstore build/spark-store-tui; \
	else \
		echo "case-insensitive build filesystem: use ./build/sparkstore"; \
	fi; \
	rm -rf "$$probe"

clean:
	rm -f *.deb
	rm -f $(SOURCE_TAR)
