clean:
	pachctl delete pipeline montage
	pachctl delete pipeline edges
	pachctl delete repo images

init:
	pachctl create repo images
	pachctl create pipeline -f pipelines/edges.pipeline.json
	pachctl create pipeline -f pipelines/montage.pipeline.json

globalid:
	pachctl put file images@master -i data/images.txt
	pachctl put file images@master -i data/images2.txt


.PHONY:
