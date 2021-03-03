.PHONY: launch-dev launch-prod

launch-dev:
	npm run mock-and-start --prefix ./backend & npm run start --prefix ./frontend

launch-prod:
	npm run build --prefix ./frontend && npm run build --prefix ./backend && npm start --prefix ./backend
