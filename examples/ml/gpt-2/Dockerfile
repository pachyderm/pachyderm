FROM nvidia/cuda:10.0-cudnn7-devel-ubuntu18.04

RUN apt-get update && \
    apt-get install -y python3-pip && \
    rm -rf /var/lib/apt/lists/*

RUN pip3 install twitterscraper gpt_2_simple tensorflow-gpu==1.14

ADD tweets.py /
ADD train.py /
ADD generate.py /
