FROM pachyderm/datascience-notebook:fde4beb9ff1afb404f0e34828adc1f311f4bf2d7 
# https://github.com/pachyderm/docker-stacks/pull/1/commits/fde4beb9ff1afb404f0e34828adc1f311f4bf2d7
ARG PACHCTL_VERSION
ENV PFS_MOUNT_DIR=/pfs

# TODO: use ARG TARGETPLATFORM to support arm builds, downloading pachctl arm64
# binary below (instead of ..._linux_amd64.tar.gz below). See:
# https://docs.docker.com/engine/reference/builder/#automatic-platform-args-in-the-global-scope

USER root
RUN apt-key adv --keyserver keyserver.ubuntu.com --recv-keys A4B469963BF863CC
RUN mkdir -p /pfs
RUN chown jovyan /pfs
RUN apt-get update && apt-get -y install curl fuse
RUN curl -f -o pachctl.tar.gz -L https://storage.googleapis.com/pachyderm-builds/pachctl_${PACHCTL_VERSION}_linux_amd64.tar.gz
RUN tar zxfv pachctl.tar.gz && mv pachctl_${PACHCTL_VERSION}_linux_amd64/pachctl /usr/local/bin/

USER $NB_UID
RUN pip install --upgrade pip

USER root
WORKDIR /app
COPY /scripts/config.sh .
RUN chmod +x config.sh

USER $NB_UID
COPY dist dist
WORKDIR /home/jovyan
RUN pip install `find /app/dist/ -name \*.whl` nbgitpuller
