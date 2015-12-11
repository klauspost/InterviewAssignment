#!/bin/sh

docker rmi tutum.co/klauspost/peytz:latest
docker build --tag=tutum.co/klauspost/peytz:latest .
echo docker push tutum.co/klauspost/peytz:latest