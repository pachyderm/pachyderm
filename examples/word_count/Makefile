wordcount:
	pachctl create repo urls
	pachctl put file urls@master -f Wikipedia
	pachctl create pipeline -f scraper.json
	pachctl create pipeline -f map/map.json
	pachctl create pipeline -f reduce.json
	pachctl flush commit reduce@master
	pachctl list file reduce@master

.PHONY:
	wordcount
