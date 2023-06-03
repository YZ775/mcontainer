
IMAGE = quay.io/cybozu/ubuntu:22.04
IMAGE_NAME = ubuntu
OS ?= $(shell uname -s)
ARCH ?= arm64
RUN_BY ?= multipass
GO_DIR = /usr/local/go/bin/go

ifeq ($(OS),Darwin)
	EXEC = multipass exec mcontainer-vm --
else
	EXEC =
endif

.PHONY: multipass-launch
multipass-launch:
	multipass launch --name mcontainer-vm --cloud-init multipass.yaml 22.04
	multipass mount $(shell pwd) mcontainer-vm:/mcontainer

.PHONY: multipass-start
multipass-start:
	multipass start mcontainer-vm

.PHONY: multipass-stop
multipass-stop:
	multipass stop mcontainer-vm

.PHONY: multipass-restart
multipass-restart:multipass-stop multipass-start

.PHONY: multipass-delete
multipass-delete:
	multipass delete mcontainer-vm
	multipass purge

.PHONY: shell
shell:
	multipass shell mcontainer-vm

.PHONY: create-rootfs
create-rootfs: config
	mkdir -p rootfs/$(IMAGE_NAME)/rootfs
	$(EXEC) sudo docker run --rm  -d --name $(IMAGE_NAME) $(IMAGE) sleep 10
	$(EXEC) sudo docker export $(IMAGE_NAME) > rootfs/$(IMAGE_NAME)/archive.tar
	tar -xvf rootfs/$(IMAGE_NAME)/archive.tar -C rootfs/$(IMAGE_NAME)/rootfs
	rm rootfs/$(IMAGE_NAME)/archive.tar

.PHONY: config
config:
	mkdir -p rootfs/$(IMAGE_NAME)
	$(EXEC) sudo runc spec
	mv config.json rootfs/$(IMAGE_NAME)

.PHONEY: clean
clean:
	rm -rf rootfs

.PHONY: build
build:
	$(EXEC) $(GO_DIR) build -o mcontainer main.go

.PHONY: run
run:
	$(EXEC) sudo $(GO_DIR) run main.go

