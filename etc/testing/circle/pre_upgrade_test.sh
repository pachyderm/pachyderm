#!/bin/bash

pachctl create repo input && 
pachctl create pipeline -f etc/testing/circle/test-pipeline.json &&
echo "foo" | pachctl put file input@master:/foo &&

# timeout for pipeline to process
sleep 15;

foo_file_count=$(pachctl list file output@master | grep foo | wc -l);
if [[ $foo_file_count != 1 ]];
then
echo "pachctl list file output@master should contain file foo but only returned: $foo_file_count";
exit 1;
fi;

if [[ $(pachctl get file output@master:/foo) != "foo" ]];
then 
echo "output@master:/foo did not contain the expected contents";
exit 1;
fi;

echo $ENT_ACT_CODE | pachctl license activate;

if [[ $(pachctl enterprise get-state | grep "ACTIVE" | wc -l) < 1 ]];
then 
echo "enterprise state expected to be ACTIVE";
exit 1;
fi;

echo "Test ran successfully!";