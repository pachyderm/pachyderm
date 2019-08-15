export BENCH_VERSION=

docker-build:
	./build/docker-build.sh
	@# Generate Dockerfile and build the supervisor into its own container.
	@# Note: if BENCH_VERSION is set, then this adds :$BENCH_VERSION to the image tag
	for bin in supervisor; do \
		image_name="pachyderm/obj-loadtest-$${bin}$${BENCH_VERSION:+:$$BENCH_VERSION}"; \
		printf "FROM ubuntu\nCOPY $${bin} /\nRUN apt-get update && apt-get install -y ca-certificates\n" >_out/Dockerfile; \
		docker build -t $${image_name} ./_out; \
		docker push $${image_name}; \
	done
	@# Not necessary, but keeps repo clean
	rm -rf ./_out
