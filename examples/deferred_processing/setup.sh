#!/bin/bash
pachctl create repo images_dp_1
pachctl create repo images_dp_2
pachctl create pipeline -f ./edges_dp.json 
pachctl create pipeline -f ./montage_dp.json 
pachctl put file images_dp_1@master -i ./images.txt
pachctl put file images_dp_1@master -i ./images2.txt
pachctl put file images_dp_2@master -i ./images3.txt

