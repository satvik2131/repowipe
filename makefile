build:
	cd frontend && npm install && npm run build
	rm -rf backend/static
	cp -r frontend/dist backend/static
	cd backend && go build -o repowipe
