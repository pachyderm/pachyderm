FROM quay.io/jupyter/scipy-notebook:lab-4.0.11

ENV PFS_MOUNT_DIR=/pfs

# TODO: use ARG TARGETPLATFORM to support arm builds, downloading pachctl arm64
# binary below (instead of ..._linux_amd64.tar.gz below). See:
# https://docs.docker.com/engine/reference/builder/#automatic-platform-args-in-the-global-scope

USER root
# The following line requires a compiled pachctl binary.
COPY ./dist-pach/pachctl/pachctl_linux_amd64_v1/pachctl /usr/local/bin/pachctl

WORKDIR /app

COPY /scripts/config.sh .
RUN chmod +x config.sh
RUN mkdir /pfs
RUN chown jovyan:users /pfs
RUN chmod 777 /pfs

USER $NB_UID
COPY dist dist
RUN pip install pip==24.0
RUN pip install `find /app/dist/ -name \*.whl` nbgitpuller==1.2.1
RUN pip install determined==0.31.0
# RUN /opt/conda/bin/jupyter lab extension disable nbclassic

WORKDIR /home/jovyan
