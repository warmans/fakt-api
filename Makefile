PROJECT_NAME=stressfaktor-api
PROJECT_VERSION=0.10.2

# Go
#-----------------------------------------------------------------------

.PHONY: test
test:
	@go test

.PHONY: build
build:
	go get
	GO15VENDOREXPERIMENT=1 go build -ldflags "-X github.com/warmans/stressfaktor-api/server.Version=$(PROJECT_VERSION)"

.PHONY: docker
docker: 
	docker build -t warmans/stressfaktor-api:$(PROJECT_VERSION) .

# Github Releases
#-----------------------------------------------------------------------

#this contains a github api token and is not included in the repo
include .make/private.mk

GH_REPO_OWNER = warmans
GH_REPO_NAME = $(PROJECT_NAME)

RELEASE_TARGET_COMMITISH = master
RELEASE_ARTIFACT_DIR = .dist
RELEASE_VERSION=$(PROJECT_VERSION)

include .make/github.mk

# Packaging
#-----------------------------------------------------------------------

PACKAGE_NAME := $(PROJECT_NAME)
PACKAGE_CONTENT_DIR := .packaging
PACKAGE_TYPE := deb
PACKAGE_OUTPUT_DIR := $(RELEASE_ARTIFACT_DIR)
PACKAGE_VERSION := $(PROJECT_VERSION)

include .make/packaging.mk

.PHONY: _configure_package
_configure_package: build
	echo "moving files into package staging area ($(PACKAGE_CONTENT_DIR))..."

	#copy binary
	@mkdir -p $(PACKAGE_CONTENT_DIR)/usr/bin/ && cp $(PROJECT_NAME) $(PACKAGE_CONTENT_DIR)/usr/bin/.

	#install config
	@install -Dm 755 init/$(PROJECT_NAME).service $(PACKAGE_CONTENT_DIR)/etc/systemd/system/$(PROJECT_NAME).service

	#setup dirs
	@mkdir -p $(PACKAGE_CONTENT_DIR)/var/lib/$(PROJECT_NAME)

# Meta
#-----------------------------------------------------------------------

.PHONY: publish
publish: build package release
