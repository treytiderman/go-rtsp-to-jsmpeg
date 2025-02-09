# BUILD: docker build -t go-ffmpeg-to-jsmpeg .
# STOP: docker stop go-ffmpeg-to-jsmpeg
# REMOVE: docker rm go-ffmpeg-to-jsmpeg
# RUN: docker run -d --name go-ffmpeg-to-jsmpeg -p 9000:9000 -v ./config:/app/config go-ffmpeg-to-jsmpeg
# EXEC: docker exec go-ffmpeg-to-jsmpeg sh

FROM golang:1.23
WORKDIR /app/src
COPY ./src /app/src
COPY ./public /app/public
COPY ./config /app/config

# Install ffmpeg
RUN apt-get update && apt-get install ffmpeg -y

# Build and run binary
RUN go build -o go-ffmpeg-to-jsmpeg
CMD ["/app/src/go-ffmpeg-to-jsmpeg"]
