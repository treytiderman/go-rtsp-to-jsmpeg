# BUILD: docker build -t go-ffmpeg-to-jsmpeg .
# STOP: docker stop go-ffmpeg-to-jsmpeg
# REMOVE: docker rm go-ffmpeg-to-jsmpeg
# RUN: docker run -d --name go-ffmpeg-to-jsmpeg -p 8000:8000 -v ./config:/app/config go-ffmpeg-to-jsmpeg

FROM golang:1.23
WORKDIR /app/src
COPY ./public /app/public
COPY ./config /app/config
COPY ./templates /app/templates

# Install ffmpeg
RUN apt-get update && apt-get install ffmpeg -y
# RUN apk update
# RUN apk add
# RUN apk add ffmpeg

# Build and run binary
RUN go build -o go-ffmpeg-to-jsmpeg
CMD ["/app/src/go-ffmpeg-to-jsmpeg"]