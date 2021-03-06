ifdef VERSION
docker_registry = n7mobile/concourse-bitbucket-pr:$(VERSION)
else
docker_registry = n7mobile/concourse-bitbucket-pr
endif

docker:
	docker build -t $(docker_registry) .

publish: docker
	docker push $(docker_registry)

test:
	go test -v ./...

run:
	./run.sh $(stage)

.DEFAULT_GOAL := docker

.PHONY: go-mod docker-build docker-push docker test