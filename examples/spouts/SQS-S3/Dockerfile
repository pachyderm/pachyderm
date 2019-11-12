FROM python:3.7.3-stretch
RUN pip install boto3
# add the spout code

ADD sqs-spout.py /sqs-spout.py
RUN chmod +x sqs-spout.py

