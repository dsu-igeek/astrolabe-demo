#!/bin/sh
docker rm astrolabe_minio
docker run --name astrolabe_minio -p 127.0.0.1:9000:9000 -v /home/dsmithuchida/astrolabe_fs_root/:/fs_root -v /home/dsmithuchida/astrolabe_repo:/astrolabe_repo astrolabe_minio  

