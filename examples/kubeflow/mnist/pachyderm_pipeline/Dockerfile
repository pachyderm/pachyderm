FROM ubuntu:19.04

RUN apt-get update -y \
    && apt-get install -y \
    curl \
    jq \
    apt-transport-https \
    ca-certificates gnupg

WORKDIR /app
ADD trigger.sh .
RUN chmod +x trigger.sh
