COMPILE_IMAGE="pachyderm/compile:$(cat etc/compile/GO_VERSION)"
COMPILE_RUN_ARGS="-d -v /var/run/socker.sock:/var/run/docker.sock --privileged=true"
VERSION_ADDITIONAL=-$(git log --pretty=format:%H | head -n 1)
LD_FLAGS="-X github.com/pachyderm/pachyderm/src/client/version.AdditionalVersion=$(VERSION_ADDITIONAL)"

TYPE=$1
NAME=$(TYPE)_compile

docker run \
  -v $(PWD):/go/src/github.com/pachyderm/pachyderm \
  -v $$GOPATH/pkg:/go/pkg \
  -v $$HOME/.cache/go-build:/root/.cache/go-build \
  --name $(NAME) \
  $(COMPILE_RUN_ARGS) $(COMPILE_IMAGE) /go/src/github.com/pachyderm/pachyderm/etc/compile/compile.sh $(TYPE) $(LD_FLAGS)