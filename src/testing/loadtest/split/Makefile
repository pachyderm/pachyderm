docker-build:
	rm -rf ./_out || true
	mkdir _out
	docker build -f Dockerfile.buildenv -t build_split_loadtest .
	@# Explanation of mounts:
	@# _out is where the binaries built in this docker image are written out for
	@# us to use:
	@#   /out <- ./_out 
	@#
	@# This re-uses your local docker cache, so that builds are faster
	@#   /root/.cache/go-build <- $HOME/.cache/go-build
	@#
	@# This makes pachyderm (including the pachyderm client and this benchmark)
	@# available to be built. Mounting all of it in (rather than e.g. copying it)
	@# means that parallel changes can be made to the benchmark, server, and
	@# client simultaneously
	@#   /go/src/.../pachyderm <- $GOPATH/src/.../pachyerm/
	@#
	@# This is necessary to build the pachyderm client (which the 'split'
	@# supervisor depends on, and which has dependencies but has no vendor
	@# directory). It accomplishes a similar goal to the mainline pachyderm's
	@# 'make docker-build' directive, which symlinks src/server/vendor into the
	@# client
	@#   /go/src/../pachyderm/src/client/vendor <- $GOPATH/.../pachyderm/src/server/vendor
	@#
	@# This avoids a stupid bug where
	@# src/client/vendor/github.com/pachyderm/pachyderm/src/client can fail to
	@# build because it (a vendored dependency) doesn't have the right
	@# dependencies to build. We prevent the go compiler from trying to build the
	@# client inside the vendored dependencies of the client by mounting a
	@# directory over it that contains no go code
	@#   /go/src/../pachyderm/src/client/vendor/github.com/pachyderm/ <- ./_out
	docker run \
		-v $${PWD}/_out:/out \
		-v $${HOME}/.cache/go-build:/root/.cache/go-build \
		-v $${GOPATH}/src/github.com/pachyderm/pachyderm:/go/src/github.com/pachyderm/pachyderm:ro \
		-v $${GOPATH}/src/github.com/pachyderm/pachyderm/src/server/vendor:/go/src/github.com/pachyderm/pachyderm/src/client/vendor:ro \
		-v $${PWD}/_out:/go/src/github.com/pachyderm/pachyderm/src/client/vendor/github.com/pachyderm:ro \
		build_split_loadtest
	# Generate Dockerfile and build the supervisor and pipeline into their own
	# containers
	for bin in supervisor pipeline; do \
		image_name="pachyderm/split-loadtest-$${bin}"; \
		printf "FROM scratch\nCOPY $${bin} /\n" >_out/Dockerfile; \
		docker build -t $${image_name} ./_out; \
		docker push $${image_name}; \
	done
	rm -rf ./_out

