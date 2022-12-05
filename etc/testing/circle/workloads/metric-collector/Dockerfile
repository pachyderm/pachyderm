FROM python:3.8-slim-buster

COPY requirements.txt requirements.txt

WORKDIR /app

COPY requirements.txt requirements.txt

RUN pip3 install -r requirements.txt

COPY . .

ENV GOOGLE_APPLICATION_CREDENTIALS=build-release-001-c3225549e514.json