setup-datums:
	-pachctl create repo data
	-pachctl create pipeline -f lots-of-datums/pipeline.json
	-mkdir -p ./lots-of-datums/data
	-node lots-of-datums/generateData.js 100 ./lots-of-datums/data
	-pachctl put file -r data@master:/ -f lots-of-datums/data
generate-datums:
	-node lots-of-datums/generateData.js 100 ./lots-of-datums/data
	-pachctl put file -r data@master:/ -f lots-of-datums/data

setup-opencv:
	-pachctl create repo images
	-pachctl put file images@master:liberty.png -f ../etc/testing/files/46Q8nDz.jpg
	-pachctl create pipeline -f https://raw.githubusercontent.com/pachyderm/pachyderm/2.4.x/examples/opencv/edges.json
	-pachctl put file images@master:AT-AT.png -f ../etc/testing/files/8MN9Kg0.jpg
	-pachctl put file images@master:kitten.png -f ../etc/testing/files/g2QnNqa.jpg
	-pachctl create pipeline -f https://raw.githubusercontent.com/pachyderm/pachyderm/2.4.x/examples/opencv/montage.json

setup-tutorial-project:
	-pachctl create project tutorial --description wow
	-pachctl create repo images --project tutorial
	-pachctl create pipeline -f https://raw.githubusercontent.com/pachyderm/pachyderm/master/examples/opencv/edges.json --project tutorial
	-pachctl put file images@master:birthday-cake.jpg -f ../etc/testing/files/FOO9q43.jpg --project tutorial
	-pachctl create pipeline -f https://raw.githubusercontent.com/pachyderm/pachyderm/master/examples/opencv/montage.json --project tutorial
	-pachctl put file images@master:pooh.jpg -f ../etc/testing/files/Togu2RY.jpg --project tutorial

setup-multi-project: setup-multi-project-1 setup-multi-project-2 setup-multi-project-cron-egress

setup-multi-project-1:
	-pachctl create project Multi-Project-A1 --description 'Contains nodes 1 and 3 of pipeline A.'
	-pachctl create project Multi-Project-A2 --description 'Contains node 2 of pipeline A.'
	-pachctl create repo images --project Multi-Project-A1
	-pachctl create pipeline -f ./multi-project/A2_edges.json --project Multi-Project-A2
	-pachctl put file images@master:birthday-cake.jpg -f ../etc/testing/files/FOO9q43.jpg --project Multi-Project-A1
	-pachctl create pipeline -f ./multi-project/A1_montage.json --project Multi-Project-A1
	-pachctl put file images@master:pooh.jpg -f ../etc/testing/files/Togu2RY.jpg --project Multi-Project-A1

setup-multi-project-2:
	-pachctl create project Multi-Project-B1 --description 'Contains nodes 1 and 4 of pipeline B.'
	-pachctl create project Multi-Project-B2 --description 'Contains nodes 2 and 3 of pipeline B.'
	-pachctl create repo images --project Multi-Project-B1
	-pachctl create pipeline -f ./multi-project/B1_edge.json --project Multi-Project-B2
	-pachctl create pipeline -f ./multi-project/B2_edge.json --project Multi-Project-B2
	-pachctl create pipeline -f ./multi-project/B3_edge.json --project Multi-Project-B1
	-pachctl put file images@master:birthday-cake.jpg -f ../etc/testing/files/FOO9q43.jpg --project Multi-Project-B1
	-pachctl put file images@master:pooh.jpg -f ../etc/testing/files/Togu2RY.jpg --project Multi-Project-B1

setup-multi-project-cron-egress:
	-pachctl create project Multi-Project-C1 --description 'Contains a cron repo.'
	-pachctl create project Multi-Project-C2 --description 'Contains a pipeline.'
	-pachctl create project Multi-Project-C3 --description 'Contains an egress repo.'
	-pachctl create pipeline -f ./multi-project/C1_cron.json --project Multi-Project-C1
	-pachctl create pipeline -f ./multi-project/C2_edges.json --project Multi-Project-C2
	-pachctl create pipeline -f ./multi-project/C3_egress.json --project Multi-Project-C3

generate-commits:
	-pachctl create repo lots-of-commits
	-node scripts/lots-of-commits/index.js 100 lots-of-commits