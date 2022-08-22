FROM ubuntu:20.04

ENV BOOTSTRAP_VERSION=1.2
ENV VERSION=1.2.23
ENV SNOWSQL_DEST=/usr/local/bin
ENV SNOWSQL_LOGIN_SHELL=/home/snowflake/.bashrc

RUN apt-get update && apt-get -y install curl unzip

RUN groupadd -r snowflake && useradd --no-log-init -r -g snowflake -d /home/snowflake -m snowflake && chown -R snowflake:snowflake /home/snowflake

WORKDIR /home/snowflake

RUN curl -O https://sfc-repo.snowflakecomputing.com/snowsql/bootstrap/${BOOTSTRAP_VERSION}/linux_x86_64/snowsql-${VERSION}-linux_x86_64.bash
RUN bash snowsql-${VERSION}-linux_x86_64.bash

USER snowflake

# snowsql attempts to write logs to ../$PWD, so run it in the context of a workspace dir where the user has write permission to the parent
WORKDIR /home/snowflake/workspace
RUN snowsql -v
