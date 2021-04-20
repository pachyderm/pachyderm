.PHONY: launch-dev launch-prod setup-auth test graphql install build ci

launch-dev:
	npm run start:dev --prefix ./backend & npm run start --prefix ./frontend

launch-prod:
	npm run build --prefix ./frontend && npm run build --prefix ./backend && npm start --prefix ./backend

launch-mock:
	npm run mock-and-start --prefix ./backend & npm run start:mock  --prefix ./frontend

setup-auth:
	npm run setup

install:
	npm install && npm install --prefix ./backend && npm install --prefix ./frontend

ci:
	npm ci && npm ci --prefix ./backend && npm ci --prefix ./frontend

build:
	npm run build --prefix ./frontend && npm run build --prefix ./backend

graphql:
	npm run generate-types --prefix ./backend && npm run generate-operations --prefix ./frontend && npm run generate-hooks --prefix ./frontend

test:
	npm test --prefix ./backend && npm test --prefix ./frontend
