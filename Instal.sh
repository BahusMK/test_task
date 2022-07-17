#!/bin/bash

mkdir ./DB_FILES 
mkdir ./pgadmin

chmod 777 ./DB_FILES
chmod 777 ./pgadmin

docker-compose build --no-cache
docker-compose up -d
