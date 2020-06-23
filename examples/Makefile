# The executable paths below are set to generic values.
# Modify them for your system by setting environment variables, using a command like this
# to both fetch the version of pachyderm you want to use and execute it.
# "../etc/fetch_release_pachctl.py 1.10.0 ; env PACHCTL=${GOPATH}/bin/pachctl make -e opencv"

SHELL := /bin/bash
PACHCTL := pachctl
KUBECTL := kubectl
# This controls which version of iris is deployed.
# acceptable values are python, julia, rstats
# invoke another flavor like this:
# env IRIS_FLAVOR=julia make -e iris
IRIS_FLAVOR := python

GATK_GERMLINE_FILES	=	gatk/GATK_Germline/data/ref/ref.dict \
				gatk/GATK_Germline/data/ref/ref.fasta \
				gatk/GATK_Germline/data/ref/ref.fasta.fai \
				gatk/GATK_Germline/data/ref/refSDF/done \
				gatk/GATK_Germline/data/ref/refSDF/format.log \
				gatk/GATK_Germline/data/ref/refSDF/mainIndex \
				gatk/GATK_Germline/data/ref/refSDF/namedata0 \
				gatk/GATK_Germline/data/ref/refSDF/nameIndex0 \
				gatk/GATK_Germline/data/ref/refSDF/namepointer0 \
				gatk/GATK_Germline/data/ref/refSDF/progress \
				gatk/GATK_Germline/data/ref/refSDF/seqdata0 \
				gatk/GATK_Germline/data/ref/refSDF/seqpointer0 \
				gatk/GATK_Germline/data/ref/refSDF/sequenceIndex0 \
				gatk/GATK_Germline/data/ref/refSDF/summary.txt \
				gatk/GATK_Germline/data/bams/father.bai \
				gatk/GATK_Germline/data/bams/father.bam \
				gatk/GATK_Germline/data/bams/mother.bai \
				gatk/GATK_Germline/data/bams/mother.bam \
				gatk/GATK_Germline/data/bams/motherICE.bai \
				gatk/GATK_Germline/data/bams/motherICE.bam \
				gatk/GATK_Germline/data/bams/motherNEX.bai \
				gatk/GATK_Germline/data/bams/motherNEX.bam \
				gatk/GATK_Germline/data/bams/motherRnaseq.bai \
				gatk/GATK_Germline/data/bams/motherRnaseq.bam \
				gatk/GATK_Germline/data/bams/motherRnaseqPP.bai \
				gatk/GATK_Germline/data/bams/motherRnaseqPP.bam \
				gatk/GATK_Germline/data/bams/son.bai \
				gatk/GATK_Germline/data/bams/son.bam

iris-base:
	$(PACHCTL) create repo training
	$(PACHCTL) create repo attributes
	$(PACHCTL) create pipeline -f ml/iris/$(IRIS_FLAVOR)_train.json 
	$(PACHCTL) create pipeline -f ml/iris/$(IRIS_FLAVOR)_infer.json

iris: iris-base
	$(PACHCTL) start transaction
	$(PACHCTL) start commit training@master
	$(PACHCTL) start commit attributes@master
	$(PACHCTL) finish transaction
	$(PACHCTL) put file training@master:iris.csv -f ml/iris/data/iris.csv
	$(PACHCTL) finish commit training@master
	$(PACHCTL) put file attributes@master:1.csv  -f ml/iris/data/test/1.csv
	$(PACHCTL) finish commit attributes@master
	$(PACHCTL) put file attributes@master:2.csv  -f ml/iris/data/test/2.csv

opencv-base:
	$(PACHCTL) create repo images
	$(PACHCTL) create pipeline -f opencv/edges.json
	$(PACHCTL) create pipeline -f opencv/montage.json

opencv: opencv-base
	$(PACHCTL) put file images@master -i opencv/images.txt
	$(PACHCTL) put file images@master -i opencv/images2.txt

opencv-delete:
	$(PACHCTL) delete pipeline montage
	$(PACHCTL) delete pipeline edges
	$(PACHCTL) delete repo images

ml/object-detection/frozen_inference_graph.pb:
	wget -p http://download.tensorflow.org/models/object_detection/ssd_mobilenet_v1_coco_11_06_2017.tar.gz --directory-prefix /tmp
	tar -C ml/object-detection --strip-components 1 -xvf /tmp/download.tensorflow.org/models/object_detection/ssd_mobilenet_v1_coco_11_06_2017.tar.gz ssd_mobilenet_v1_coco_11_06_2017/frozen_inference_graph.pb

# This target is not intended to be invoked directly, it's intended for other targets to be built on.
# OBJECT_IMAGES is set to no operation by default, so this pipeline can use the images repo of
# opencv when used for testing.  If invoked from the intended object-detection target, below,
# a separate images repo will be created for object-detection.
OBJECT_MODEL_PIPELINE = $(PACHCTL) create pipeline -f ml/object-detection/model.json
OBJECT_DETECT_PIPELINE = $(PACHCTL) create pipeline -f ml/object-detection/detect.json
object-detection-base: 
	@$(OBJECT_IMAGES)
	$(PACHCTL) create repo training
	@$(OBJECT_MODEL_PIPELINE)
	@$(OBJECT_DETECT_PIPELINE)

object-detection-data: ml/object-detection/frozen_inference_graph.pb
	$(PACHCTL) put file training@master:frozen_inference_graph.pb -f ml/object-detection/frozen_inference_graph.pb
	$(PACHCTL) put file images@master:dogs.jpg -f ml/object-detection/images/dogs.jpg

object-detection: OBJECT_IMAGES = $(PACHCTL) create repo images
object-detection:  object-detection-base object-detection-data

object-detection-delete: object-detection-delete
	-$(PACHCTL) delete pipeline detect
	-$(PACHCTL) delete pipeline model
	-$(PACHCTL) delete repo training
	-$(PACHCTL) delete repo images

hyperparameter-common:
	$(PACHCTL) create repo raw_data
	$(PACHCTL) create repo parameters

hyperparameter-base: hyperparameter-common
	$(PACHCTL) create pipeline -f ml/hyperparameter/split.json
	$(PACHCTL) create pipeline -f ml/hyperparameter/model.json
	$(PACHCTL) create pipeline -f ml/hyperparameter/test.json
	$(PACHCTL) create pipeline -f ml/hyperparameter/select.json

hyperparameter-data: 
	$(PACHCTL) start transaction
	$(PACHCTL) start commit raw_data@master
	$(PACHCTL) start commit parameters@master
	$(PACHCTL) finish transaction
	$(PACHCTL) put file raw_data@master:iris.csv -f ml/hyperparameter/data/noisy_iris.csv
	$(PACHCTL) put file parameters@master:c_parameters.txt -f ml/hyperparameter/data/parameters/c_parameters.txt --split line --target-file-datums 1
	$(PACHCTL) put file parameters@master:gamma_parameters.txt -f ml/hyperparameter/data/parameters/gamma_parameters.txt --split line --target-file-datums 1
	$(PACHCTL) finish commit raw_data@master
	$(PACHCTL) finish commit parameters@master

hyperparameter: hyperparameter-common hyperparameter-base hyperparameter-data

hyperparameter-delete:
	-$(PACHCTL) delete pipeline select
	-$(PACHCTL) delete pipeline test
	-$(PACHCTL) delete pipeline split
	-$(PACHCTL) delete pipeline model
	-$(PACHCTL) delete repo parameters
	-$(PACHCTL) delete repo raw_data

$(GATK_GERMLINE_FILES): 
	mkdir -p gatk/GATK_Germline
	wget -p https://s3-us-west-1.amazonaws.com/pachyderm.io/Examples_Data_Repo/GATK_Germline.zip --directory-prefix=/tmp
	unzip -o /tmp/s3-us-west-1.amazonaws.com/pachyderm.io/Examples_Data_Repo/GATK_Germline.zip  data/ref/ref.dict data/ref/ref.fasta data/ref/ref.fasta.fai data/ref/refSDF/* data/bams/*  -d gatk/GATK_Germline

gatk-base: 
	$(PACHCTL) create repo reference
	$(PACHCTL) create repo samples
	$(PACHCTL) create pipeline -f gatk/likelihoods.json
	$(PACHCTL) create pipeline -f gatk/joint-call.json

gatk-data: $(GATK_GERMLINE_FILES)
	$(PACHCTL) start transaction
	$(PACHCTL) start commit reference@master
	$(PACHCTL) start commit samples@master
	$(PACHCTL) finish transaction
	$(PACHCTL) put file reference@master:ref.dict  -f gatk/GATK_Germline/data/ref/ref.dict
	$(PACHCTL) put file reference@master:ref.fasta -f gatk/GATK_Germline/data/ref/ref.fasta
	$(PACHCTL) put file reference@master:ref.fasta.fai -f gatk/GATK_Germline/data/ref/ref.fasta.fai
	$(PACHCTL) put file reference@master:refSDF -r -f gatk/GATK_Germline/data/ref/refSDF
	$(PACHCTL) put file samples@master:mother/mother.bam -f gatk/GATK_Germline/data/bams/mother.bam
	$(PACHCTL) put file samples@master:mother/mother.bai -f gatk/GATK_Germline/data/bams/mother.bai
	$(PACHCTL) put file samples@master:father/father.bam -f gatk/GATK_Germline/data/bams/father.bam
	$(PACHCTL) put file samples@master:father/father.bai -f gatk/GATK_Germline/data/bams/father.bai	
	$(PACHCTL) finish commit reference@master
	$(PACHCTL) finish commit samples@master

gatk: gatk-base gatk-data

gatk-delete:
	-$(PACHCTL) delete pipeline joint_call
	-$(PACHCTL) delete pipeline likelihoods
	-$(PACHCTL) delete repo reference
	-$(PACHCTL) delete repo samples

delete:
	yes | $(PACHCTL) delete all

clean: delete
	rm -fr gatk/GATK_Germline
	rm -f  ml/object-detection/frozen_inference_graph.pb

minikube:echo hi
	minikube start
	@WHEEL="-\|/"; \
	until minikube ip 2>/dev/null; do \
	    WHEEL=$${WHEEL:1}$${WHEEL:0:1}; \
	    echo -en "\e[G\e[K$${WHEEL:0:1} waiting for minikube to start..."; \
	    sleep 1; \
	done
	$(PACHCTL) deploy local
	@until "$$(
		$(KUBECTL) get po \
		  -l suite=pachyderm,app=dash \
		  -o jsonpath='{.items[0].status.conditions[?(@.type=="Ready")].status}')" = True; \
	do \
		WHEEL=$${WHEEL:1}$${WHEEL:0:1}; \
		$(ECHO) -en "\e[G\e[K$${WHEEL:0:1} waiting for pachyderm to start..."; \
		sleep 1; \
	done
	pachctl port-forward &
