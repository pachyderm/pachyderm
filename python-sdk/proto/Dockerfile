FROM python:3.8-slim

RUN apt-get update -y && apt-get install -y git

RUN python3 -m pip install 'grpcio-tools==1.51.3'
RUN python3 -m pip install  'betterproto[compiler] @ git+https://github.com/pachyderm/python-betterproto.git@bonenfant/INT-1159'

COPY entrypoint.sh /bin
ENTRYPOINT ["/bin/entrypoint.sh"]
WORKDIR /work
