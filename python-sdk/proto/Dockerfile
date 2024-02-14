FROM python:3.8-slim

RUN apt-get update -y && apt-get install git -y
RUN python3 -m pip install 'black==24.1.1' 'grpcio-tools==1.51.3'
RUN python3 -m pip install 'betterproto[compiler] @ git+https://github.com/pachyderm/python-betterproto.git@remove-generated-servicer'

COPY entrypoint.sh /bin
ENTRYPOINT ["/bin/entrypoint.sh"]
WORKDIR /work
