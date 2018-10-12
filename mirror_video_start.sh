#!/bin/bash

# raspivid starting command for the mirror

/usr/bin/raspivid -w 160 -h 160 -r /home/pi/video_pipe -x /home/pi/motion_vectors -rf rgb -t 0 -n -o /dev/null -br 70 -co 25 -v -rot 270 &
GODEBUG=cgocheck=0 /home/pi/go/bin/mirror2 --addr=":8080" --graph tolbert_faces.graph --video /home/pi/video_pipe --motion /home/pi/motion_vectors --mbx 10 --mby 10 --magnitude 1 --totalMotion 3
