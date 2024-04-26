FROM python:3.8-slim

RUN apt-get update -y && apt-get install git -y
RUN python3 -m pip install 'black==24.1.1' 'grpcio-tools==1.51.3'
RUN python3 -m pip install 'betterproto[compiler] @ https://github.com/pachyderm/python-betterproto/releases/download/v2.0.0b6%2Bpachv1.0.4/betterproto-2.0.0b6+pachv1.0.4-py3-none-any.whl'

COPY entrypoint.sh /bin
ENTRYPOINT ["/bin/entrypoint.sh"]
WORKDIR /work
