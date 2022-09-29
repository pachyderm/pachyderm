#!/bin/bash
pachctl create repo images
pachctl create repo edges
pachctl create repo montage

echo "data" | pachctl put file images@master:/file1
echo "data" | pachctl put file images@master:/file2
echo "data" | pachctl put file images@dev:/file1
echo "data" | pachctl put file images@dev:/file2

echo "data" | pachctl put file edges@master:/file1
echo "data" | pachctl put file edges@master:/file2
echo "data" | pachctl put file edges@dev:/file1
echo "data" | pachctl put file edges@dev:/file2

echo "data" | pachctl put file montage@master:/file1
echo "data" | pachctl put file montage@master:/file2
echo "data" | pachctl put file montage@dev:/file1
echo "data" | pachctl put file montage@dev:/file2
