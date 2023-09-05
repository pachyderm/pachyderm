setup-datums:
	-pachctl create repo data
	-pachctl create pipeline -f lots-of-datums/pipeline.pipeline.json
	-mkdir -p ./lots-of-datums/data
	-node lots-of-datums/generateData.js 100 ./lots-of-datums/data
	-pachctl put file -r data@master:/ -f lots-of-datums/data
generate-datums:
	-node lots-of-datums/generateData.js 100 ./lots-of-datums/data
	-pachctl put file -r data@master:/ -f lots-of-datums/data

setup-opencv:
	-pachctl create repo images
	-pachctl put file images@master:liberty.png -f ../etc/testing/files/46Q8nDz.jpg
	-pachctl create pipeline -f https://raw.githubusercontent.com/pachyderm/pachyderm/master/examples/opencv/edges.pipeline.json
	-pachctl put file images@master:AT-AT.png -f ../etc/testing/files/8MN9Kg0.jpg
	-pachctl put file images@master:kitten.png -f ../etc/testing/files/g2QnNqa.jpg
	-pachctl create pipeline -f https://raw.githubusercontent.com/pachyderm/pachyderm/master/examples/opencv/montage.pipeline.json

setup-tutorial-project:
	-pachctl create project tutorial --description wow
	-pachctl create repo images --project tutorial
	-pachctl create pipeline -f https://raw.githubusercontent.com/pachyderm/pachyderm/master/examples/opencv/edges.pipeline.json --project tutorial
	-pachctl put file images@master:birthday-cake.jpg -f ../etc/testing/files/FOO9q43.jpg --project tutorial
	-pachctl create pipeline -f https://raw.githubusercontent.com/pachyderm/pachyderm/master/examples/opencv/montage.pipeline.json --project tutorial
	-pachctl put file images@master:pooh.jpg -f ../etc/testing/files/Togu2RY.jpg --project tutorial

setup-multi-project: setup-multi-project-1 setup-multi-project-2 setup-multi-project-cron-egress

setup-multi-project-1:
	-pachctl create project Multi-Project-A1 --description 'Contains nodes 1 and 3 of pipeline A.'
	-pachctl create project Multi-Project-A2 --description 'Contains node 2 of pipeline A.'
	-pachctl create repo images --project Multi-Project-A1
	-pachctl create pipeline -f ./multi-project/A2_edges.pipeline.json --project Multi-Project-A2
	-pachctl put file images@master:birthday-cake.jpg -f ../etc/testing/files/FOO9q43.jpg --project Multi-Project-A1
	-pachctl create pipeline -f ./multi-project/A1_montage.pipeline.json --project Multi-Project-A1
	-pachctl put file images@master:pooh.jpg -f ../etc/testing/files/Togu2RY.jpg --project Multi-Project-A1

setup-multi-project-2:
	-pachctl create project Multi-Project-B1 --description 'Contains nodes 1 and 4 of pipeline B.'
	-pachctl create project Multi-Project-B2 --description 'Contains nodes 2 and 3 of pipeline B.'
	-pachctl create repo images --project Multi-Project-B1
	-pachctl create pipeline -f ./multi-project/B1_edge.pipeline.json --project Multi-Project-B2
	-pachctl create pipeline -f ./multi-project/B2_edge.pipeline.json --project Multi-Project-B2
	-pachctl create pipeline -f ./multi-project/B3_edge.pipeline.json --project Multi-Project-B1
	-pachctl put file images@master:birthday-cake.jpg -f ../etc/testing/files/FOO9q43.jpg --project Multi-Project-B1
	-pachctl put file images@master:pooh.jpg -f ../etc/testing/files/Togu2RY.jpg --project Multi-Project-B1

setup-multi-project-cron-egress:
	-pachctl create project Multi-Project-C1 --description 'Contains a cron repo.'
	-pachctl create project Multi-Project-C2 --description 'Contains a pipeline.'
	-pachctl create project Multi-Project-C3 --description 'Contains an egress repo.'
	-pachctl create pipeline -f ./multi-project/C1_cron.pipeline.json --project Multi-Project-C1
	-pachctl create pipeline -f ./multi-project/C2_edges.pipeline.json --project Multi-Project-C2
	-pachctl create pipeline -f ./multi-project/C3_egress.pipeline.json --project Multi-Project-C3

generate-commits:
	-pachctl create repo lots-of-commits
	-node scripts/lots-of-commits/index.js 100 lots-of-commits

failing-pipeline:
	-pachctl create project Pipeline-Failure --description 'Contains a single pipeline that will fail because the docker image is incorrect.'
	-pachctl create repo images --project Pipeline-Failure
	-echo '{"pipeline":{"name":"edges","project":{"name":"Pipeline-Failure"}},"description": "A broken pipeline.","input":{"pfs":{"glob": "/*","repo": "images"}},"transform":{"cmd":["python3","/edges.py"],"image":"pachyderm/doesnotexist:1.0"}}' | pachctl create pipeline

# RBAC
rbac-setup:
	-echo "pizza" | pachctl auth use-auth-token
	-pachctl auth set cluster none user:kilgore@kilgore.trout

rbac-root:
	-echo "pizza" | pachctl auth use-auth-token
	-pachctl create project root --description "Project made by pach root. You can see nothing."
	-pachctl create repo repo1 --project root
	-echo '{"pipeline":{"name":"pipeline1","project":{"name":"root"}},"description":"A sleepy pipeline.","input":{"pfs":{"glob":"/*","repo":"repo1","project":"root"}},"transform":{"cmd":["sh"],"stdin": ["sleep 0"]}}' | pachctl create pipeline

rbac-tomcat:
	-echo "pizza" | pachctl auth use-auth-token
	-pachctl auth get-robot-token tomcat --quiet | pachctl auth use-auth-token
	-pachctl create project tomcat --description "allClusterUsers have projectOwner on the project. There is one repo."
	-pachctl create repo images --project tomcat
	-pachctl auth set project tomcat projectOwner allClusterUsers

rbac-viper:
	-echo "pizza" | pachctl auth use-auth-token
	-pachctl auth whoami
	-pachctl auth get-robot-token viper --quiet | pachctl auth use-auth-token
	-pachctl create project viper --description "Kilgore has projectViewer. Kilgore has repoReader on one of three repos and repoOwner on another"
	-pachctl create repo repo1 --project viper
	-pachctl create repo repo2 --project viper
	-pachctl create repo repo3 --project viper
	-pachctl auth set project viper projectViewer user:kilgore@kilgore.trout
	-pachctl auth set repo repo1 repoReader user:kilgore@kilgore.trout --project viper
	-pachctl auth set repo repo3 repoOwner user:kilgore@kilgore.trout --project viper

rbac-warthog:
	-echo "pizza" | pachctl auth use-auth-token
	-pachctl auth whoami
	-pachctl auth get-robot-token warthog --quiet | pachctl auth use-auth-token
	-pachctl create project warthog --description "Kilgore has repoReader on one of the two repos."
	-pachctl create repo hidden --project warthog
	-pachctl create repo visible --project warthog
	-pachctl auth set repo visible repoReader user:kilgore@kilgore.trout --project warthog

generate-rbac: rbac-setup rbac-root rbac-tomcat rbac-viper rbac-warthog
	-echo "pizza" | pachctl auth use-auth-token
	
setup-service-pipeline:
	-pachctl create repo service-input
	-pachctl put file service-input@master:birthday-cake.jpg -f ../etc/testing/files/FOO9q43.jpg
	-pachctl create pipeline -f ./pipelines/simple-service.pipeline.json

setup-spout-pipeline:
	-pachctl create pipeline -f https://raw.githubusercontent.com/pachyderm/pachyderm/master/examples/spouts/spout101/pipelines/spout.pipeline.json
	-pachctl create pipeline -f https://raw.githubusercontent.com/pachyderm/pachyderm/master/examples/spouts/spout101/pipelines/processor.pipeline.json

setup-spout-service-pipeline:
	-pachctl create pipeline -f ./pipelines/spout-and-service.pipeline.json
	-pachctl create pipeline -f https://raw.githubusercontent.com/pachyderm/pachyderm/master/examples/spouts/spout101/pipelines/processor.pipeline.json

cross-montage:
	-pachctl create repo images
	-pachctl create pipeline -f ./cross-montage-pipelines/edges-a.pipeline.json
	-pachctl create pipeline -f ./cross-montage-pipelines/edges-b.pipeline.json
	-pachctl create pipeline -f ./cross-montage-pipelines/cross-montage.pipeline.json
	