FROM python:3.7-slim-stretch
RUN mkdir /app 
ADD ./producer.py requirements.txt /app/
WORKDIR /app
RUN pip install -r requirements.txt
CMD ["/app/producer.py"]
