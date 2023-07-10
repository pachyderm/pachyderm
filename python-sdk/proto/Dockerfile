FROM python:3.8-slim

RUN python3 -m pip install 'grpcio-tools==1.51.3'
RUN python3 -m pip install 'betterproto[compiler] @ https://github.com/pachyderm/python-betterproto/releases/download/v2.0.0b5%2Bpachv1.0.2/betterproto-2.0.0b5+pachv1.0.2-py3-none-any.whl'

COPY entrypoint.sh /bin
ENTRYPOINT ["/bin/entrypoint.sh"]
WORKDIR /work
