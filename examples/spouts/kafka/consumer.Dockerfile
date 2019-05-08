FROM python:3.7-slim-stretch
RUN mkdir /app 
ADD ./consumer.py requirements.txt /app/
WORKDIR /app
RUN pip install -r requirements.txt
CMD ["/app/consumer.py"]
