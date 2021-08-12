#!/bin/bash

# build
docker build -t vladalv/http-server-error:v1 .

# upload
docker push vladalv/http-server-error:v1