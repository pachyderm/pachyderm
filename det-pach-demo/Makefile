



install:
	pachctl create repo input;
	pachctl put file input@master:/hello.txt
	pachctl create pipeline det -f det-pipeline.json
	pachctl put file goodbye.txt

clean:
	echo "yes" | pachctl delete --all
