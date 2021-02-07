wordcount:
	pachctl create repo urls
	cd data && pachctl put file urls@master -f Wikipedia
	pachctl create pipeline -f pipelines/scraper.json
	pachctl create pipeline -f pipelines/map.json
	pachctl create pipeline -f pipelines/reduce.json

clean:
	pachctl delete pipeline reduce
	pachctl delete pipeline map
	pachctl delete pipeline scraper
	pachctl delete repo urls
	pachctl delete repo scraper	
	pachctl delete repo map	
	pachctl delete repo reduce


.PHONY:
	wordcount
