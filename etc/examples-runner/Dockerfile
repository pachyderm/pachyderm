FROM pachyderm/pachctl
RUN apt update -y && apt install -y ca-certificates git jq
WORKDIR /app
RUN git clone https://github.com/pachyderm/pachyderm.git /app
COPY run.sh .
ENTRYPOINT ["./run.sh"]
