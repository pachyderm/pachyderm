.PHONY: opencv object_detection delete minikube

opencv:
	pachctl create-repo images
	pachctl create-pipeline -f opencv/edges.json
	pachctl create-pipeline -f opencv/montage.json
	pachctl put-file images master -i opencv/images.txt
	pachctl put-file images master -i opencv/images2.txt

object_detection:
	pachctl create-repo images
	pachctl create-repo training
	pachctl create-pipeline -f ml/object-detection/model.json
	pachctl create-pipeline -f ml/object-detection/detect.json
	pachctl put-file training master frozen_inference_graph.pb -f ~/ssd_mobilenet_v1_coco_11_06_2017/frozen_inference_graph.pb
	pachctl put-file images master dogs.jpg -f ml/object-detection/images/dogs.jpg
	pachctl put-file images master kites.jpg -f ml/object-detection/images/kites.jpg


delete:
	yes | pachctl delete-all

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
	