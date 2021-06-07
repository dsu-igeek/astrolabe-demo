#!/bin/sh
docker rm astrolabe_server
docker run --name astrolabe_server -p 127.0.0.1:1323:1323 -v /home/dsmithuchida/astrolabe_fs_root/:/fs_root -v /home/dsmithuchida/astrolabe_repo:/astrolabe_repo astrolabe_server  

