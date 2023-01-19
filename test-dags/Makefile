setup-datums:
	-pachctl create pipeline -f lots-of-datums/pipeline.json
	-pachctl create repo data
	-mkdir -p ./lots-of-datums/data
	-node lots-of-datums/generateData.js 100 ./lots-of-datums/data
	-pachctl put file -r data@master:/ -f lots-of-datums/data
generate-datums:
	-node lots-of-datums/generateData.js 100 ./lots-of-datums/data
	-pachctl put file -r data@master:/ -f lots-of-datums/data

setup-opencv:
	-pachctl create repo images
	-pachctl put file images@master:liberty.png -f http://imgur.com/46Q8nDz.png
	-pachctl create pipeline -f https://raw.githubusercontent.com/pachyderm/pachyderm/2.4.x/examples/opencv/edges.json
	-pachctl put file images@master:AT-AT.png -f http://imgur.com/8MN9Kg0.png
	-pachctl put file images@master:kitten.png -f http://imgur.com/g2QnNqa.png
	-pachctl create pipeline -f https://raw.githubusercontent.com/pachyderm/pachyderm/2.4.x/examples/opencv/montage.json

