#!/bin/bash

# raspivid starting command for the mirror

/usr/bin/raspivid -w 160 -h 160 -r /home/pi/video_pipe -x /home/pi/motion_vectors -rf rgb -t 0 -n -o /dev/null -v -roi 0.2134,0.2638,0.3354,0.4464 -ISO 0 -ss 500000 &
/home/pi/go/bin/mirror2 --addr=":8080" --graph /home/pi/tolbert_faces_201812.graph --video /home/pi/video_pipe --motion /home/pi/motion_vectors --mbx 10 --mby 10 --magnitude 1 --totalMotion 3
