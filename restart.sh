#!/bin/bash
echo "\n---------- Restarting nodes ----------\n"
docker compose down -v

echo "\n---------- Removing old data ----------"
rm -rf /home/kshitij/raft-data/node-1
rm -rf /home/kshitij/raft-data/node-2
rm -rf /home/kshitij/raft-data/node-3
rm -rf /home/kshitij/raft-data/node-4

echo "\n---------- Creating Folders for Volumes ----------"
mkdir /home/kshitij/raft-data/node-1
mkdir /home/kshitij/raft-data/node-2
mkdir /home/kshitij/raft-data/node-3
mkdir /home/kshitij/raft-data/node-4


echo "\n---------- Building images ----------"
docker compose build --no-cache

echo "\n---------- Starting nodes ----------"
docker compose up -d