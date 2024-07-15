FROM node:16-buster-slim

WORKDIR /usr/src/app

RUN apt-get update && apt-get install

ADD package.json /package.json
ADD index.js /index.js

run npm i
