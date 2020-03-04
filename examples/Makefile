.PHONY: opencv object_detection delete minikube gatk hyperparameter

JQ := jq
JQFLAGS :=
DOCKER := docker
ECHO := echo
TOUCH := touch
CP := cp
RM := rm
STATS := .enable_stats=true
DETECTMODEL := .pipeline.name="object-detect"|.input.cross[1].pfs.repo="object-model"|.input.cross[1].pfs.name="model"
HYPERPSPLIT := .input.cross[0].pfs.repo="hp-split"|.input.cross[0].pfs.name="split"
HYPERPMODEL := .input.cross[1].pfs.repo="hp-model"|.input.cross[1].pfs.name="model"
HYPERPTEST := .input.cross[0].pfs.repo="hp-test"|.input.cross[0].pfs.name="test"


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

opencv:
	pachctl create repo images
	pachctl create pipeline -f opencv/edges.json
	pachctl create pipeline -f opencv/montage.json
	pachctl put file images@master -i opencv/images.txt
	pachctl put file images@master -i opencv/images2.txt

opencv-testing: 
	pachctl create repo images
	$(JQ) "$(STATS)" opencv/edges.json | pachctl create pipeline -f -
	$(JQ) "$(STATS)" opencv/montage.json | pachctl create pipeline -f -
	pachctl start transaction
	pachctl start commit images@master
	pachctl finish transaction
	pachctl put file images@master -i opencv/images.txt
	pachctl put file images@master -i opencv/images2.txt
	pachctl put file images@master -i opencv/images3.txt
	pachctl put file images@master -i opencv/images4.txt
	pachctl delete file images@master:/w7RVTsv.jpg
	pachctl finish commit images@master

opencv-testing-delete:
	-pachctl delete pipeline montage
	-pachctl delete pipeline edges

opencv-delete: opencv-testing-delete
	-pachctl delete repo images

ml/object-detection/frozen_inference_graph.pb:
	wget -p http://download.tensorflow.org/models/object_detection/ssd_mobilenet_v1_coco_11_06_2017.tar.gz --directory-prefix /tmp
	tar -C ml/object-detection --strip-components 1 -xvf /tmp/download.tensorflow.org/models/object_detection/ssd_mobilenet_v1_coco_11_06_2017.tar.gz ssd_mobilenet_v1_coco_11_06_2017/frozen_inference_graph.pb


object-detection-images-repo: ml/object-detection/frozen_inference_graph.pb
	pachctl create repo images

object-detection-no-images-repo: ml/object-detection/frozen_inference_graph.pb

object-detection-demo: object-detection-no-images-repo
	pachctl create repo training
	pachctl put file training@master:frozen_inference_graph.pb -f ml/object-detection/frozen_inference_graph.pb
	pachctl put file images@master:dogs.jpg -f ml/object-detection/images/dogs.jpg

object-detection:  object-detection-images-repo object-detection-demo
	pachctl create pipeline -f ml/object-detection/model.json
	pachctl create pipeline -f ml/object-detection/detect.json

object-detection-testing:  object-detection-demo
	$(JQ) '$(STATS)|.pipeline.name="object-model"'  ml/object-detection/model.json | pachctl create pipeline -f -
	$(JQ) '$(STATS)|$(DETECTMODEL)'  ml/object-detection/detect.json | pachctl create pipeline -f -
	pachctl put file images@master:airplane.jpg -f ml/object-detection/images/airplane.jpg

object-detection-testing-delete:
	-pachctl delete pipeline object-detect
	-pachctl delete pipeline object-model
	-pachctl delete repo training

object-detection-delete: object-detection-delete
	-pachctl delete pipeline detect
	-pachctl delete pipeline model
	-pachctl delete repo training
	-pachctl delete repo images

hyperparameter-common:
	pachctl create repo raw_data
	pachctl create repo parameters
	pachctl put file raw_data@master:iris.csv -f ml/hyperparameter/data/noisy_iris.csv
	pachctl put file parameters@master:c_parameters.txt -f ml/hyperparameter/data/parameters/c_parameters.txt --split line --target-file-datums 1
	pachctl put file parameters@master:gamma_parameters.txt -f ml/hyperparameter/data/parameters/gamma_parameters.txt --split line --target-file-datums 1

hyperparameter: hyperparameter-common
	pachctl create pipeline -f ml/hyperparameter/split.json
	pachctl create pipeline -f ml/hyperparameter/model.json
	pachctl create pipeline -f ml/hyperparameter/test.json
	pachctl create pipeline -f ml/hyperparameter/select.json

hyperparameter-testing: hyperparameter-common
	$(JQ) '$(STATS)|.pipeline.name="hp-split"' ml/hyperparameter/split.json | pachctl create pipeline -f -
	$(JQ) '$(STATS)|.pipeline.name="hp-model"' ml/hyperparameter/model.json | pachctl create pipeline -f -
	$(JQ) '$(STATS)|.pipeline.name="hp-test"|$(HYPERPSPLIT)|$(HYPERPMODEL)' ml/hyperparameter/test.json | pachctl create pipeline -f -
	$(JQ) '$(STATS)|.pipeline.name="hp-select"|$(HYPERPTEST)|$(HYPERPMODEL)'  ml/hyperparameter/select.json | pachctl create pipeline -f -

hyperparameter-testing-delete:
	-pachctl delete pipeline hp-select
	-pachctl delete pipeline hp-test
	-pachctl delete pipeline hp-split
	-pachctl delete pipeline hp-model
	-pachctl delete repo parameters
	-pachctl delete repo raw_data

hyperparameter-delete:
	-pachctl delete pipeline select
	-pachctl delete pipeline test
	-pachctl delete pipeline split
	-pachctl delete pipeline model
	-pachctl delete repo parameters
	-pachctl delete repo raw_data

$(GATK_GERMLINE_FILES): 
	mkdir -p gatk/GATK_Germline
	wget -p https://s3-us-west-1.amazonaws.com/pachyderm.io/Examples_Data_Repo/GATK_Germline.zip --directory-prefix=/tmp
	unzip -o /tmp/s3-us-west-1.amazonaws.com/pachyderm.io/Examples_Data_Repo/GATK_Germline.zip  data/ref/ref.dict data/ref/ref.fasta data/ref/ref.fasta.fai data/ref/refSDF/* data/bams/*  -d gatk/GATK_Germline

gatk: $(GATK_GERMLINE_FILES)
	pachctl create repo reference
	pachctl put file reference@master:ref.dict  -f gatk/GATK_Germline/data/ref/ref.dict
	pachctl put file reference@master:ref.fasta -f gatk/GATK_Germline/data/ref/ref.fasta
	pachctl put file reference@master:ref.fasta.fai -f gatk/GATK_Germline/data/ref/ref.fasta.fai
	pachctl put file reference@master:refSDF -r -f gatk/GATK_Germline/data/ref/refSDF
	pachctl create repo samples
	pachctl start commit samples@master
	pachctl put file samples@master:mother/mother.bam -f gatk/GATK_Germline/data/bams/mother.bam
	pachctl put file samples@master:mother/mother.bai -f gatk/GATK_Germline/data/bams/mother.bai
	pachctl finish commit samples@master
	pachctl start commit samples@master
	pachctl put file samples@master:father/father.bam -f gatk/GATK_Germline/data/bams/father.bam
	pachctl put file samples@master:father/father.bai -f gatk/GATK_Germline/data/bams/father.bai	
	pachctl finish commit samples@master
	pachctl create pipeline -f gatk/likelihoods.json
	pachctl create pipeline -f gatk/joint-call.json

gatk-testing: gatk
	pachctl start commit samples@master
	pachctl put file samples@master:son/son.bam -f gatk/GATK_Germline/data/bams/son.bam
	pachctl put file samples@master:son/son.bai -f gatk/GATK_Germline/data/bams/son.bai	
	pachctl finish commit samples@master

gatk-delete:
	-pachctl delete pipeline joint_call
	-pachctl delete pipeline likelihoods
	-pachctl delete repo reference
	-pachctl delete repo samples

gatk-testing-delete: gatk-delete

testing-delete: gatk-testing-delete hyperparameter-testing-delete object-detection-testing-delete opencv-testing-delete
	-pachctl delete repo images

testing: opencv-testing object-detection-testing hyperparameter-testing gatk-testing

delete:
	yes | pachctl delete all

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
	pachctl deploy local
	@until "$$(
		kc get po \
		  -l suite=pachyderm,app=dash \
		  -o jsonpath='{.items[0].status.conditions[?(@.type=="Ready")].status}')" = True; \
	do \
		WHEEL=$${WHEEL:1}$${WHEEL:0:1}; \
		echo -en "\e[G\e[K$${WHEEL:0:1} waiting for pachyderm to start..."; \
		sleep 1; \
	done
	pachctl port-forward &
