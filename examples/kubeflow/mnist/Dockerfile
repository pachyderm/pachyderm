FROM python:3
RUN pip3 install kfp
WORKDIR /app
ADD trigger.py .
ADD version.json .
RUN chmod +x trigger.py
