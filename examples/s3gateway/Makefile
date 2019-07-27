.PHONY: image pipeline tf_job_image push_tf_job_image
IMAGE_TAG = 7

image: 
	docker build -t papposilenus/pach_minio_copy:$(IMAGE_TAG) -f ./Dockerfile.pach_minio_copy .
        docker push papposilenus/pach_minio_copy:$(IMAGE_TAG)

tf_job_image: 
	docker build -t papposilenus/tf_job_s3_gateway:$(IMAGE_TAG) -f ./Dockerfile.tf_job_s3_gateway .

push_tf_job_image:
        docker push papposilenus/tf_job_s3_gateway:$(IMAGE_TAG)

pipeline: image
	sed s/TAG/$(IMAGE_TAG)/g < pach_minio_copy.json.template > pach_minio_copy.json
	pachctl update pipeline -f pach_minio_copy.json --reprocess
