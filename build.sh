#!/bin/sh

userID=$(id -u)
groupID=$(id -g)

mkdir -p ./build

echo "Building Docker image"
imageID=$(docker build -f docker/Dockerfile --build-arg UID="${userID}" --build-arg GID="${groupID}" -q .)

echo "Running build"
docker run --rm -v ./build:/build "$imageID"

echo "Removing image"
docker image rm "$imageID"
