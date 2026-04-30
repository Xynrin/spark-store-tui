PKGROOT := package-root
DEB := spark-store-tui_0.7.2-1_all.deb
SOURCE_TAR := spark-store-tui-deb-source-0.7.2.tar.gz
SOURCE_DIR := spark-store-tui-deb-source-0.7.2

.PHONY: all check build source clean local-install

all: check build

check:
	bash -n $(PKGROOT)/usr/bin/spark-store-tui
	@if command -v shellcheck >/dev/null 2>&1; then shellcheck $(PKGROOT)/usr/bin/spark-store-tui; else echo "shellcheck not found; skipped"; fi

build:
	(cd $(PKGROOT) && find usr -type f -print0 | sort -z | xargs -0 md5sum > DEBIAN/md5sums)
	dpkg-deb --build --root-owner-group $(PKGROOT) $(DEB)

source:
	tar --exclude='.git' --exclude='*.deb' --exclude='*.tar.gz' --exclude='.cache' --exclude='build' --exclude='dist' --exclude='package-root/tmp' --transform='s,^,$(SOURCE_DIR)/,' -czf $(SOURCE_TAR) .

local-install: build
	sudo apt install ./$(DEB)

clean:
	rm -f $(DEB)
	rm -f $(SOURCE_TAR)
	rm -f $(PKGROOT)/DEBIAN/md5sums
