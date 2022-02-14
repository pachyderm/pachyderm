# Set as you wish
DOCKER_ACCOUNT := pachyderm
CONTAINER_NAME := example-group
CONTAINER_VERSION := 2.1.0
CONTAINER_TAG := $(DOCKER_ACCOUNT)/$(CONTAINER_NAME):$(CONTAINER_VERSION)

# Build Tag Push a new image to Docker Hub
docker-image-group-retail:
	@docker build -t $(CONTAINER_TAG) .  --platform linux/amd64
	@docker push $(CONTAINER_TAG)

# This target can be used during the development/update of the pipeline's code of the retail example. 
# Change your python's file (check the `src` dir of this example) or update your template, then run the following target.
# Your code will be built, tagged, and your pipeline updated all at once, using Minikube's local registry.
# Make sure that you have run `make setup-retail` and `make deploy-retail` first to create your mockup data and repos.
# Then run `dev-update-image-group-retail` every time your source code has updates.
# This target uses the pipeline template retail_group.jsonnet.
dev-update-image-group-retail:
	$(eval IMAGE_SHA=$(shell docker build -t ${CONTAINER_TAG} --platform=linux/amd64 -q .))
	$(eval IMAGE_TAG=$(shell echo $(IMAGE_SHA) |cut -c 8-20))
	docker tag $(CONTAINER_TAG) $(CONTAINER_TAG).$(IMAGE_TAG)
	pachctl update pipeline --jsonnet pipelines/retail/retail_group.jsonnet --arg tag=$(CONTAINER_VERSION).$(IMAGE_TAG) --arg suffix=jsonnet

# Lab-1 Create mockup data for lab example  
setup-lab:
	mkdir -p ./labresults 
	printf "PID|PATID1\nORC|ORD777888\nOBX|1|NM|2093-3^Cholesterol|196|mg/dL\nOBX|2|NM|2571-8^Triglyceride|100|mg/dL" \
	>./labresults/T1606707557-LIPID-PATID1-CLIA24D9871327.txt
	printf "PID|PATID2\nORC|ORD777889\nOBX|1|NM|2093-3^Cholesterol|212|mg/dL\nOBX|2|NM|2571-8^Triglyceride|110|mg/dL" \
	>./labresults/T1606331395-LIPID-PATID2-CLIA24D9871327.txt
	printf "PID|PATID3\nORC|ORD777890\nOBX|1|NM|2093-3^Cholesterol|188|mg/dL\nOBX|2|NM|2571-8^Triglyceride|89|mg/dL" \
	>./labresults/T1606707579-LIPID-PATID3-CLIA24D9871327.txt
	printf "PID|PATID4\nORC|ORD777891\nOBX|1|NM|2093-3^Cholesterol|230|mg/dL\nOBX|2|NM|2571-8^Triglyceride|132|mg/dL" \
	>./labresults/T1606707597-LIPID-PATID4-CLIA24D9871327.txt
	printf "PID|PATID1\nORC|ORD777892\nOBX|1|NM|2093-3^Cholesterol|199|mg/dL\nOBX|2|NM|2571-8^Triglyceride|101|mg/dL" \
	>./labresults/T1606707613-LIPID-PATID1-CLIA24D9871328.txt
	printf "PID|PATID3\nORC|ORD777893\nOBX|1|NM|2093-3^Cholesterol|187|mg/dL\nOBX|2|NM|2571-8^Triglyceride|92|mg/dL" \
	>./labresults/T1606707635-LIPID-PATID3-CLIA24D9871328.txt

# Lab-2 Populate initial repos with mock up data
deploy-lab:
	pachctl create repo labresults
	pachctl put file -r labresults@master:/ -f labresults
	pachctl list file labresults@master

# Lab-3 Create pipelines
create-lab:
	pachctl create pipeline -f pipelines/lab/lab_group_by_hospital.json 
	pachctl create pipeline -f pipelines/lab/lab_group_by_patient.json
	pachctl create pipeline -f pipelines/lab/reduce_group_by_hospital.json 
	pachctl create pipeline -f pipelines/lab/reduce_group_by_patient.json

# Retail-1 Create mockup data for retail example
setup-retail:
	mkdir -p ./stores ./purchases ./returns
	echo "Creating stores..."
	printf '{"storeid":"1", "name":"washington st.","address":{"zipcode":"02108","country":"US"}}' \
	>./stores/STOREID1.txt
	printf '{"storeid":"2", "name":"hollywood bvd.","address":{"zipcode":"02108","country":"US"}}' \
	>./stores/STOREID2.txt
	printf '{"storeid":"3", "name":"beverly hills","address":{"zipcode":"90210","country":"US"}}' \
	>./stores/STOREID3.txt
	printf '{"storeid":"4", "name":"mariposa st.","address":{"zipcode":"94107","country":"US"}}' \
	>./stores/STOREID4.txt
	printf '{"storeid":"5", "name":"tremont st.","address":{"zipcode":"90210","country":"US"}}' \
	>./stores/STOREID5.txt
	echo "Creating purchases..."
	printf "ORDER_HEADER|W080520|2020-10-10\nSKU|1001|QTY|2|UNIT_PRICE|100.00\n" \
	 >./purchases/ORDERW080520_STOREID1.txt
	printf "ORDER_HEADER|W080521|2020-10-10\nSKU|2002|QTY|1|UNIT_PRICE|250.00\n" \
	 >./purchases/ORDERW080521_STOREID1.txt
	printf "ORDER_HEADER|W078929|2020-10-10\nSKU|0615|QTY|1|UNIT_PRICE|75.00\n" \
	 >./purchases/ORDERW078929_STOREID2.txt
	printf "ORDER_HEADER|W598471|2020-10-10\nSKU|1001|QTY|3|UNIT_PRICE|100.00\n" \
	 >./purchases/ORDERW598471_STOREID3.txt
	printf "ORDER_HEADER|W080231|2020-10-10\nSKU|1001|QTY|2|UNIT_PRICE|100.00\n" \
	 >./purchases/ORDERW080231_STOREID5.txt
	printf "ORDER_HEADER|W080528|2020-10-10\nSKU|2002|QTY|1|UNIT_PRICE|250.00\n" \
	 >./purchases/ORDERW080528_STOREID5.txt
	echo "Creating returns..."
	printf "ORDER_HEADER|W080520|2020-10-12\nSKU|1001|QTY|1|UNIT_PRICE|100.00\n" \
	 >./returns/ORDERW080520_STOREID1.txt
	printf "ORDER_HEADER|W080231|2020-10-10\nSKU|1001|QTY|1|UNIT_PRICE|100.00\n" \
	 >./returns/ORDERW080231_STOREID5.txt
	printf "ORDER_HEADER|W080528|2020-10-10\nSKU|2002|QTY|1|UNIT_PRICE|250.00\n" \
	 >./returns/ORDERW080528_STOREID5.txt
	printf "ORDER_HEADER|W261452|2020-10-10\nSKU|0246|QTY|1|UNIT_PRICE|50.00\n" \
	 >./returns/ORDERW261452_STOREID0.txt

# Retail-2 Populate initial repos with mock up data
deploy-retail:
	pachctl create repo stores
	pachctl create repo purchases
	pachctl create repo returns
	pachctl put file -r stores@master:/ -f stores
	pachctl list file stores@master
	pachctl put file -r purchases@master:/ -f purchases
	pachctl list file purchases@master	
	pachctl put file -r returns@master:/ -f returns
	pachctl list file returns@master	

# Retail-3 Create pipelines
create-pipeline-retail:
	pachctl create pipeline -f pipelines/retail/retail_group.json





