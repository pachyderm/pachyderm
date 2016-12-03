build:
	docker build -t pachyderm_jupyter .

launch: build
	pachctl version
	pachctl create-repo foo
	echo 'hai' | pachctl put-file foo master -c file
	pachctl create-job -f jupyter.json

clean-launch:
	pachctl delete-repo foo


