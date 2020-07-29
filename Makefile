IMAGE_NAME=di-velocity
GIT_SHA = $(shell git rev-parse HEAD)
IMAGE_REPO=jasonblanchard/${IMAGE_NAME}
LOCAL_TAG = ${IMAGE_REPO}
LATEST_TAG= ${IMAGE_REPO}:latest
SHA_TAG = ${IMAGE_REPO}:${GIT_SHA}

createdb:
	# createuser -e -d -P -E di
	createdb -U di -e -O di di_velocity

dropdb:
	dropdb di_velocity

dbmigrate:
	migrate -database postgres://di:di@localhost:5432/di_velocity?sslmode=disable -path db/migrations up

build:
	docker build -t ${LOCAL_TAG} .

tag: build
	docker tag ${LOCAL_TAG} ${SHA_TAG}

push: tag
	docker push ${LATEST_TAG}
	docker push ${SHA_TAG}

build_migrations:
	docker build -t ${LOCAL_TAG}-migrations -f db/Dockerfile .

tag_migrations: build_migrations
	docker tag ${LOCAL_TAG}-migrations ${LOCAL_TAG}-migrations:${GIT_SHA}

push_migrations: tag_migrations
	docker push ${LOCAL_TAG}-migrations
	docker push ${LOCAL_TAG}-migrations:${GIT_SHA}
