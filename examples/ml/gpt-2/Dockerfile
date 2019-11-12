FROM tensorflow/tensorflow:1.14.0-gpu-py3

RUN apt-get update && \
    apt-get install -y python3-pip && \
    rm -rf /var/lib/apt/lists/*

RUN pip3 install twitterscraper gpt_2_simple

ADD tweets.py /
ADD train.py /
ADD generate.py /
