setup:
	rm -rf ./pfs
	mkdir -p ./pfs/readings/ID1234 ./pfs/parameters ./pfs/out
	printf "1\n2\n3\n4\n5" >./pfs/readings/ID1234/file1.txt
	printf "1\n2\n3\n4\n5" >./pfs/readings/ID1234/file2.txt
	printf "1\n2\n3\n4\n5" >./pfs/readings/ID1234/file3.txt
	printf "1\n2\n3\n4\n5" >./pfs/readings/ID1234/file4.txt
	printf "1\n2\n3\n4\n5" >./pfs/readings/ID1234/file5.txt
	printf "1\n2\n3\n4\n5" >./pfs/readings/ID1234/file6.txt
	printf "1\n2\n3\n4\n5" >./pfs/parameters/file1.txt
	printf "1\n2\n3\n4\n5" >./pfs/parameters/file2.txt
	printf "1\n2\n3\n4\n5" >./pfs/parameters/file3.txt
	printf "1\n2\n3\n4\n5" >./pfs/parameters/file4.txt
	printf "1\n2\n3\n4\n5" >./pfs/parameters/file5.txt
	printf "1\n2\n3\n4\n5" >./pfs/parameters/file6.txt
	docker build -t joins-example .

run:
	docker run -v ${PWD}/pfs:/pfs joins-example
