#!/bin/sh

userID=$(id -u)
groupID=$(id -g)

mkdir -p ./build

echo "Building Docker image & Running build"
imageID=$(docker build -f docker/Dockerfile --build-arg UID="${userID}" --build-arg GID="${groupID}" --target builder -q .)

echo "Copying output"
docker run --rm -v ./build:/target "$imageID" cp /build/docker_launcher /target

echo "Removing image"
docker image rm "$imageID"
