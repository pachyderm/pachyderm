build:
	docker build -t pachyderm_jupyter .

prep: build
	pachctl version
	./citibike_data/load.sh
	./weather_data/load.sh
	pachctl create-pipeline -f ./sales_data/pipeline.json
	pachctl create-job -f jupyter.json

launch-jupyter:
	pachctl create-job -f jupyter.json

clean-launch:
	pachctl delete-repo trips
	pachctl delete-repo weather
	pachctl delete-repo sales


