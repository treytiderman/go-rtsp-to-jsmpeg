# FFmpeg Commands

Stream input to Stdout as mpegts

```sh
ffmpeg -stream_loop -1 -re -i video.mp4 -f mpegts -codec:v mpeg1video -

ffmpeg -stream_loop -1 -re -i video.mp4 -f mpegts -codec:v mpeg1video -vf scale=640:-1 -codec:a mp2 -ar 44100 -ac 1 -b:a 128k -r 24 -b 0 -

ffmpeg -i rtsp://user:password@192.168.1.22:554/cam/realmonitor?channel=1&subtype=0 -f mpegts -codec:v mpeg1video -

ffmpeg -rtsp_transport udp -i rtsp://user:password@192.168.1.22:554/cam/realmonitor?channel=1&subtype=0 -f mpegts -codec:v mpeg1video -r 30 -

ffmpeg -rtsp_transport tcp -i rtsp://user:password@192.168.1.22:554/cam/realmonitor?channel=1&subtype=0 -f mpegts -codec:v mpeg1video -r 30 -

ffmpeg -f dshow -i "video=HD Web Camera" -f mpegts -codec:v mpeg1video -r 30 -

ffmpeg -f gdigrab -framerate 10 -offset_x 0 -offset_y 0 -video_size 3840x2160 -show_region 1 -i desktop -f mpegts -codec:v mpeg1video -r 30 -
```

Get USB devices available Windows

```sh
ffmpeg -list_devices true -f dshow -i dummy
```

