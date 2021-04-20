FROM node:lts-alpine3.13

WORKDIR /usr/src/app

ADD . .

RUN apk add --no-cache \
    build-base \
    g++ \
    cairo-dev \
    jpeg-dev \
    pango-dev \
    giflib-dev

RUN make ci
RUN make build

CMD ["npm", "start", "--prefix", "./backend"]
EXPOSE 4000
