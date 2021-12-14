#!/bin/sh
docker rm astrolabe_repo_server
docker run --name astrolabe_repo_server -p 127.0.0.1:1323:1323 \
-v `pwd`/conf/s3_config.json:/etc/astrolabe_conf/s3config.json \
-it \
astrolabe_repo_server

# old -v options to docker
#-v /home/dsmithuchida/astrolabe_fs_root/:/fs_root \
#-v /home/dsmithuchida/astrolabe_repo:/astrolabe_repo \
#-v `pwd`/conf/pes:/etc/astrolabe_conf/pes \
