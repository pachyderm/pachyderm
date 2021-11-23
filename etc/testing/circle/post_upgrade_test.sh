#!/bin/bash

echo "bar" | pachctl put file input@master:/bar &&

sleep 5;

if [[ $(pachctl get file output@master:/bar) != "bar" ]];
then 
echo "output@master:/bar did not contain the expected contents"
exit 1;
fi

# check the previous file is still get-able
if [[ $(pachctl get file output@master:/foo) != "foo" ]];
then 
echo "output@master:/foo did not contain the expected contents"
exit 1;
fi

if [[ $(pachctl enterprise get-state | grep "ACTIVE" | wc -l) < 1 ]];
then 
echo "enterprise state expected to be ACTIVE";
exit 1;
fi

echo "Test ran successfully!"