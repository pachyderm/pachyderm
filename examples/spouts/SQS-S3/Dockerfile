FROM python:3.7.3-stretch

# add spout client
ADD sqs-spout.py /sqs-spout.py
ADD requirements.txt /requirements.txt
RUN pip install --no-cache-dir -r /requirements.txt


