node_modules/:
	npm install

dist/: node_modules/
	mage -v
	npm run build

dev: dist/
	docker compose up -d --build

force: 
	docker compose down
	mage -v
	npm run build
	docker compose up -d --build
	
.PHONY: dev force

