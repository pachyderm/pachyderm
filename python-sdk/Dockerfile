FROM python:3.8-slim

RUN apt-get update && apt-get install -y gcc git

RUN pip install 'poetry==1.1.8'
COPY pyproject.toml poetry.lock ./
RUN poetry install

COPY entrypoint.sh /bin
ENTRYPOINT ["/bin/entrypoint.sh"]
WORKDIR /work
