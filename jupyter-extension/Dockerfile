FROM jupyter/scipy-notebook:lab-3.3.3
# https://github.com/pachyderm/docker-stacks/pull/1/commits/fde4beb9ff1afb404f0e34828adc1f311f4bf2d7

ENV PFS_MOUNT_DIR=/pfs

# TODO: use ARG TARGETPLATFORM to support arm builds, downloading pachctl arm64
# binary below (instead of ..._linux_amd64.tar.gz below). See:
# https://docs.docker.com/engine/reference/builder/#automatic-platform-args-in-the-global-scope

USER root
COPY ./dist-pach/pachctl/pachctl_linux_amd64_v1/pachctl /usr/local/bin/pachctlhctl /usr/local/bin/

USER $NB_UID
RUN pip install --upgrade pip

USER root
WORKDIR /app
COPY /scripts/config.sh .
RUN chmod +x config.sh
RUN mkdir /pfs
RUN chown jovyan:users /pfs
RUN chmod 777 /pfs

USER $NB_UID
COPY dist dist
WORKDIR /home/jovyan
RUN pip install `find /app/dist/ -name \*.whl` nbgitpuller
RUN pip install determined==0.23.3
RUN /opt/conda/bin/jupyter lab extension disable nbclassic
