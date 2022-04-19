PREFIX			  ?= registry.cn-hangzhou.aliyuncs.com/rdc-incubator
TAG				  ?= $(shell date +%s)
SHADOW_IMAGE	  =  kt-connect-shadow
SHADOW_BASE_IMAGE =  shadow-base
ROUTER_IMAGE	  =  kt-connect-router

# run mod tidy
mod:
	go mod tidy -compat=1.17

# run unit test
test:
	mkdir -p artifacts/report/coverage
	go test -v -cover -coverprofile c.out.tmp ./...
	cat c.out.tmp | grep -v "_mock.go" > c.out
	go tool cover -html=c.out -o artifacts/report/coverage/index.html

# build kt project
compile:
	goreleaser --snapshot --skip-publish --rm-dist

# check the style
check:
	go vet ./pkg/... ./cmd/...

# build ktctl
ktctl:
	GOARCH=amd64 GOOS=linux go build -o artifacts/ktctl/ktctl-linux ./cmd/ktctl
	GOARCH=amd64 GOOS=darwin go build -o artifacts/ktctl/ktctl-darwin ./cmd/ktctl
	GOARCH=amd64 GOOS=windows go build -o artifacts/ktctl/ktctl-windows ./cmd/ktctl

# build this image before shadow
shadow-base:
	docker build -t $(PREFIX)/$(SHADOW_BASE_IMAGE):$(TAG) -f build/docker/shadow/Dockerfile_base .

# build shadow
shadow:
	GOARCH=amd64 GOOS=linux go build -gcflags "all=-N -l" -o artifacts/shadow/shadow-linux-amd64 cmd/shadow/main.go
	docker build -t $(PREFIX)/$(SHADOW_IMAGE):$(TAG) -f build/docker/shadow/Dockerfile .

# shadow
shadow-local:
	go build -gcflags "all=-N -l" -o artifacts/shadow/shadow-local cmd/shadow/main.go

# dlv for debug
shadow-dlv:
	make shadow TAG=latest
	scripts/build-shadow-dlv

# build router
router:
	GOARCH=amd64 GOOS=linux go build -gcflags "all=-N -l" -o artifacts/router/router-linux-amd64 cmd/router/main.go
	docker build -t $(PREFIX)/$(ROUTER_IMAGE):$(TAG) -f build/docker/router/Dockerfile .

# clean up workspace
clean:
	rm -fr artifacts dist
