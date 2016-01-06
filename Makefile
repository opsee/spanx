all: fmt build

build:
	gb build

clean:
	rm -fr target bin pkg

fmt:
	@gofmt -w ./

deps:
	./deps.sh

migrate:
	migrate -url $(POSTGRES_CONN) -path ./migrations up

docker: fmt
	docker run \
		--link postgresql:postgresql \
		--env-file ./$(APPENV) \
		-e "TARGETS=linux/amd64" \
		-v `pwd`:/build quay.io/opsee/build-go \
		&& docker build -t quay.io/opsee/spanx .

run: docker
	docker run \
		--link postgresql:postgresql \
		--env-file ./$(APPENV) \
		-e AWS_DEFAULT_REGION \
		-e AWS_ACCESS_KEY_ID \
		-e AWS_SECRET_ACCESS_KEY \
		-p 9095:9095 \
		--rm \
		quay.io/opsee/spanx:latest

.PHONY: docker run migrate clean all
