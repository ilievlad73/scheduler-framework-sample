#!/bin/bash

# build
docker build -t vladalv/http-server:v1 .

# upload
docker push vladalv/http-server:v1