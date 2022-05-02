up:
	docker-compose -f docker-compose.yml up -d --build

down:
	docker-compose -f docker-compose.yml down --volumes

test:
	docker-compose -f docker-compose.yml up -d --build
	-go test ./... -v
	# docker-compose -f docker-compose.yml down --volumes
