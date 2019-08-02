.PHONY: image pipeline tf_job_image push_tf_job_image
IMAGE_TAG = 1.9.2

tf_job_s3_gateway_minio_image: 
	docker build -t pachyderm/tf_job_s3_gateway_minio:$(IMAGE_TAG) -f ./Dockerfile.tf_job_s3_gateway_minio .

tf_job_s3_gateway_minio_registry: 
	docker push pachyderm/tf_job_s3_gateway_minio:$(IMAGE_TAG)

tf_job_s3_gateway_image: 
	docker build -t pachyderm/tf_job_s3_gateway:$(IMAGE_TAG) -f ./Dockerfile.tf_job_s3_gateway .

tf_job_s3_gateway_registry:
	docker push pachyderm/tf_job_s3_gateway:$(IMAGE_TAG)

