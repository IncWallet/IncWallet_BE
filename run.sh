#!/bin/sh bash

ps aux | grep '[r]un.sh' | awk '{ print $2}' | grep -v "^$$\$" | xargs kill -9
docker rm -f incwallet
docker network create --driver bridge incwallet_net || true
docker pull isyyyy/incwallet
docker run -ti --restart=always --net incwallet_net --name incwallet -d -p 9000:9000 isyyyy/incwallet
docker run --net incwallet_net --name mongo -d -p 27018:27017 mongo
shopt -s expand_aliases
alias wic='docker exec -it incwallet wic'