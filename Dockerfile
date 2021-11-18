FROM jupyter/base-notebook

USER root
RUN apt-get update && apt-get -y install curl fuse
RUN curl -f -o pachctl.deb -L https://github.com/pachyderm/pachyderm/releases/download/v2.0.1/pachctl_2.0.1_amd64.deb && dpkg -i pachctl.deb

USER $NB_UID

# Install in the default python3 environment
RUN pip install --quiet --no-cache-dir 'jupyterlab==3.2.3' && \
    fix-permissions "${CONDA_DIR}" && \
    fix-permissions "/home/${NB_USER}"


RUN mkdir /home/$NB_USER/myextension

WORKDIR /home/$NB_USER/myextension
COPY --chown=$NB_UESR jupyterlab_pachyderm ./jupyterlab_pachyderm
COPY --chown=$NB_UESR schema ./schema
COPY --chown=$NB_UESR src ./src
COPY --chown=$NB_UESR style ./style

COPY --chown=$NB_UESR install.json \
     package.json \
     pyproject.toml \
     README.md \
     setup.cfg \
     setup.py \
     tsconfig.json \
     webpack.config.js \
     ./


RUN pip install -e .
RUN jupyter labextension develop . --overwrite
RUN jupyter server extension enable jupyterlab_pachyderm
RUN npm install
RUN npm run build

WORKDIR /home/$NB_USER
RUN echo '{"pachd_address": "grpc://host.docker.internal:30650", "source": 2}' | pachctl config set context "local" --overwrite && pachctl config set active-context "local"
