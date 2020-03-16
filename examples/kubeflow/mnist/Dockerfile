FROM python:3
RUN pip3 install kfp
WORKDIR /app
ADD pipeline.py .
RUN chmod +x pipeline.py
