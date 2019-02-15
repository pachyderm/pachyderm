export BENCH_VERSION=

docker-build:
	./build/docker-build.sh
	@# Generate Dockerfile and build the supervisor and pipeline into their own
	@# containers
	@# Note: if BENCH_VERSION is set, then this adds :$BENCH_VERSION to the image tag
	for bin in supervisor pipeline; do \
		image_name="pachyderm/split-loadtest-$${bin}$${BENCH_VERSION:+:$$BENCH_VERSION}"; \
		printf "FROM scratch\nCOPY $${bin} /\n" >_out/Dockerfile; \
		docker build -t $${image_name} ./_out; \
		docker push $${image_name}; \
	done
	@# Not necessary, but keeps repo clean
	rm -rf ./_out

