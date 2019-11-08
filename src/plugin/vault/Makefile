VERSION_ADDITIONAL = -$(shell git log --pretty=format:%H | head -n 1)
LD_FLAGS = -X github.com/pachyderm/pachyderm/src/plugin/vault/vendor/github.com/pachyderm/pachyderm/src/client/version.AdditionalVersion=$(VERSION_ADDITIONAL)

plugin:
	GO111MODULE=on go install -mod=vendor -ldflags "$(LD_FLAGS)" ./pachyderm-plugin

test:
	./etc/start-vault.sh
	./etc/setup-vault.sh
	go test -v -count=1 .

.PHONY: \
	test
