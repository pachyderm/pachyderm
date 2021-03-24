.PHONY: launch-dev launch-prod setup-auth

launch-dev:
	npm run mock-and-start --prefix ./backend & npm run start --prefix ./frontend

launch-prod:
	npm run build --prefix ./frontend && npm run build --prefix ./backend && npm start --prefix ./backend

setup-auth:
	npm run setup

install:
	npm install && npm install --prefix ./backend && npm install --prefix ./frontend

graphql:
	npm run generate-types --prefix ./backend && npm run generate-operations --prefix ./frontend
