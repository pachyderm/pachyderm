FROM tensorflow/tensorflow:1.15.0-gpu-py3

RUN apt-get update && \
    apt-get install -y python3-pip && \
    rm -rf /var/lib/apt/lists/*

RUN pip3 install twitter_scraper gpt_2_simple bpython

ADD tweets.py /
ADD train.py /
ADD generate.py /
