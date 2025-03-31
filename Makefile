# Set your Docker Hub username and image name
DOCKER_USER = rasmusraiha
IMAGE_NAME = uno-bot
VERSION = latest  # You can change this to a specific version if needed


dev:
	@go run .

build:
	@docker build -t $(DOCKER_USER)/$(IMAGE_NAME):$(VERSION) .

push:
	@docker push $(DOCKER_USER)/$(IMAGE_NAME):$(VERSION)

# Build and push in one command
deploy: build push

test:
	@docker-compose up --force-recreate --remove-orphans