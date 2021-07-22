FROM python:3.9
RUN pip3 install kfp
WORKDIR /app
ADD pipeline.py .
RUN chmod +x pipeline.py
