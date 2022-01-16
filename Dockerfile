FROM pachyderm/datascience-notebook:fde4beb9ff1afb404f0e34828adc1f311f4bf2d7 
# https://github.com/pachyderm/docker-stacks/pull/1/commits/fde4beb9ff1afb404f0e34828adc1f311f4bf2d7
ARG PACHCTL_VERSION
USER root
RUN mkdir -p /pfs
RUN chown jovyan /pfs
RUN apt-get update && apt-get -y install curl fuse
RUN curl -f -o pachctl.deb -L https://github.com/pachyderm/pachyderm/releases/download/v${PACHCTL_VERSION}/pachctl_${PACHCTL_VERSION}_amd64.deb
RUN dpkg -i pachctl.deb

USER $NB_UID
RUN pip install --upgrade pip

USER root
WORKDIR /app
COPY /scripts/config.sh .
RUN chmod +x config.sh

USER $NB_UID
COPY dist dist
WORKDIR /home/jovyan
RUN pip install /app/dist/jupyterlab_pachyderm-0.1.0b3-py3-none-any.whl jupyterlab_pachyderm_theme==0.1.2 nbgitpuller
