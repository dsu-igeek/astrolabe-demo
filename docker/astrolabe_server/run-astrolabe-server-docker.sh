#!/bin/sh
docker rm astrolabe_server
docker run --name astrolabe_server -p 127.0.0.1:1323:1323 \
-v /home/dsmithuchida/astrolabe_fs_root/:/fs_root \
-v /home/dsmithuchida/astrolabe_repo:/astrolabe_repo \
-v `pwd`/conf/pes:/etc/astrolabe_conf/pes \
-v `pwd`/conf/s3config.json:/etc/astrolabe_conf/s3config.json \
-it \
dsmithuchida/astrolabe_server:Jun-11-2021-10-45-15

