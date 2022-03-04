FROM node:16-buster-slim

WORKDIR /usr/src/app

ADD . .

RUN apt-get update && apt-get install -y python3 make pkg-config \
    build-essential libcairo2-dev libpango1.0-dev libjpeg-dev libgif-dev librsvg2-dev

# update with new canvas versions
RUN npm install canvas@2.8.0

RUN make ci

RUN make build

RUN make prune-deps

CMD ["npm", "start", "--prefix", "./backend"]
EXPOSE 4000
