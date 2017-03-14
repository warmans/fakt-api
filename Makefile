PROJECT_OWNER=warmans
PROJECT_NAME=fakt-api
PROJECT_VERSION=3.1.0
DOCKER_NAME=$(PROJECT_OWNER)/$(PROJECT_NAME)

# Go
#-----------------------------------------------------------------------

.PHONY: test
test:
	go test ./server/...

.PHONY: build
build:
	GO15VENDOREXPERIMENT=1 \
	GOOS=linux \
	go build -ldflags "-X github.com/warmans/fakt-api/server.Version=$(PROJECT_VERSION)" -o .build/$(PROJECT_NAME)

.PHONY: build-cli
build-cli:
	go build -o .build/fakt-maint github.com/warmans/fakt-api/cmd/maint

# Packaging
#-----------------------------------------------------------------------

.PHONY: docker-build
docker-build:
	docker build -t $(DOCKER_NAME):$(PROJECT_VERSION) .

.PHONY: docker-publish
docker-publish:
	docker push $(DOCKER_NAME):$(PROJECT_VERSION)


# Meta
#-----------------------------------------------------------------------

.PHONY: publish
publish: test build docker-build docker-publish
