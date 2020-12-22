FROM python:3

WORKDIR /app
ADD ./src/requirements.txt requirements.txt
RUN pip install --no-cache-dir -r requirements.txt

ADD ./src/consumer/main.py consumer/main.py
ADD ./src/processor/main.py processor/main.py