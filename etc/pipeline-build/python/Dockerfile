FROM python:3

LABEL name="Pachyderm" \
      vendor="Pachyderm"

COPY LICENSE /licenses

RUN apt update && apt install -y libgl1-mesa-glx

WORKDIR /app
ADD etc/pipeline-build/python/build.sh .
ADD etc/pipeline-build/python/run.sh .
