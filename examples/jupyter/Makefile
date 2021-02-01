v=4

build:
	
	# Build your image
	docker image build -t pachyderm-notebook:$(v) .
	# Tag your image -> repl
	docker tag pachyderm-notebook:$(v) npepin/pachyderm-notebook:$(v)
	# Push your image on Docker hub
	docker push npepin/pachyderm-notebook:$(v)

