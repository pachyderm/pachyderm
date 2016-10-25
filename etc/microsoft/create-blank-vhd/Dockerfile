FROM python:2.7.12-alpine

ADD . /src
WORKDIR /src

RUN apk add --update \
    build-base \
    libffi-dev \
    openssl-dev \
  && pip install -r requirements.txt

ENTRYPOINT ["python", "create_blank_vhd.py"]

