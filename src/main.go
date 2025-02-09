package main

import (
	"log"
	"log/slog"
	"os"
	"time"
)

func main() {

	// Set up structured logging | slog.NewJSONHandler or slog.NewTextHandler
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})))

	// Load Stream Config File ./config/streams.json
	getStreamConfigFile()

	// Start streams
	slog.Info("get streams", "streams", getStreams())
	for _, stream := range getStreams() {
		time.Sleep(100 * time.Millisecond)
		startStream(stream.Id)
	}

	// Start the HTTP server
	err := httpServerStart()
	log.Fatal(err)

}
