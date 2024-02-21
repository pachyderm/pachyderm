FROM node:20.11-buster-slim

WORKDIR /usr/src/app

RUN apt-get update && apt-get install -y \
    # required for frontend dependency canvas
    build-essential libcairo2-dev libpango1.0-dev libjpeg-dev libgif-dev librsvg2-dev \
    && rm -rf /var/lib/apt/lists/*

ADD . .

RUN make docker-ci \
    && make build \
    && make clean-deps

WORKDIR /usr/src/app/backend
# needed to run npm start
RUN npm i module-alias


FROM node:20.11-buster-slim

ARG DOCKER_TAG=${DOCKER_TAG:-local}
ENV REACT_APP_RELEASE_VERSION=${DOCKER_TAG:-local}

WORKDIR /usr/src/app

COPY --from=0 /usr /usr

CMD ["npm", "start", "--prefix", "./backend"]
EXPOSE 4000
