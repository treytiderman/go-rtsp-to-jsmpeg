package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"os/exec"
	"regexp"
	"time"
)

type Stream struct {
	Id     int64  `json:"id"`
	Name   string `json:"name"`
	FFmpeg string `json:"ffmpeg"`
	Status string `json:"status"`

	cmd *exec.Cmd
}

var streamConfigs = []*Stream{}

// Helpers

func generateId() int64 {
	time.Sleep(1 * time.Millisecond)
	return time.Now().UnixMilli()
}

func stringToCommand(input string) []string {
	re := regexp.MustCompile(`"([^"]*)"|(\S+)`)
	matches := re.FindAllStringSubmatch(input, -1)
	var output []string
	for _, match := range matches {
		if match[1] != "" {
			output = append(output, match[1])
		} else if match[2] != "" {
			output = append(output, match[2])
		}
	}
	return output
}

func saveStreamConfigFile() {
	slog.Info("save stream config file", "count", len(streamConfigs))

	file, err := os.Create("../config/streams.json")
	if err != nil {
		slog.Error("failed to create config file", "error", err)
		return
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "    ")
	err = encoder.Encode(streamConfigs)
	if err != nil {
		slog.Error("failed to encode stream config file", "error", err)
		return
	}

}

func getStreamConfigFile() {
	slog.Info("get stream config file")

	// Read File
	file, err := os.Open("../config/streams.json")
	if err != nil {
		slog.Error("failed to open config file", "error", err)
		return
	}
	defer file.Close()

	// Decode JSON
	var tempStreamConfigs = []Stream{}
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&tempStreamConfigs)
	if err != nil {
		slog.Error("failed to decode stream config file", "error", err)
	}

	// Create New Streams from config
	for _, stream := range tempStreamConfigs {
		newStreamWithId(stream.Name, stream.FFmpeg, stream.Id)
	}
}

// Streams

func newStream(name string, ffmpeg string) *Stream {
	split := stringToCommand(ffmpeg)

	// Create a new stream
	stream := Stream{
		Id:     generateId(),
		Name:   name,
		FFmpeg: ffmpeg,
		Status: "stopped",
		cmd:    exec.Command(split[0], split[1:]...),
	}
	slog.Info("new stream created", "id", stream.Id, "name", stream.Name, "ffmpeg", stream.FFmpeg)

	// Save the stream
	streamConfigs = append(streamConfigs, &stream)
	saveStreamConfigFile()

	return &stream
}

func newStreamWithId(name string, ffmpeg string, id int64) Stream {
	split := stringToCommand(ffmpeg)

	// Create a new stream
	stream := Stream{
		Id:     id,
		Name:   name,
		FFmpeg: ffmpeg,
		Status: "stopped",
		cmd:    exec.Command(split[0], split[1:]...),
	}
	slog.Info("new stream created", "id", stream.Id, "name", stream.Name, "ffmpeg", stream.FFmpeg)

	// Save the stream
	streamConfigs = append(streamConfigs, &stream)
	saveStreamConfigFile()

	return stream
}

func getStreamById(id int64) *Stream {
	for _, stream := range streamConfigs {
		if stream.Id == id {
			return stream
		}
	}
	return &Stream{}
}

func getStreamByName(name string) *Stream {
	for _, stream := range streamConfigs {
		if stream.Name == name {
			return stream
		}
	}
	return &Stream{}
}

func getStreams() []*Stream {
	return streamConfigs
}

func updateStreamName(id int64, name string) {
	slog.Info("update stream name", "id", id, "name", name)
	stream := getStreamById(id)
	stream.Name = name
	saveStreamConfigFile()
}

func removeStream(id int64) {
	slog.Info("remove stream", "id", id, "stream", getStreamById(id))
	for i, stream := range streamConfigs {
		if stream.Id == id {

			// Stop stream if needed
			if stream.Status == "running" {
				stopStream(id)
			}

			// Remove config
			streamConfigs = append(streamConfigs[:i], streamConfigs[i+1:]...)
			saveStreamConfigFile()
			return

		}
	}
}

func clearAllStreams() {
	slog.Info("clear all streams")
	for _, stream := range getStreams() {
		stopStream(stream.Id)
	}
	streamConfigs = []*Stream{}
	saveStreamConfigFile()
}

func startStream(id int64) {
	stream := getStreamById(id)

	// Return if already started
	if stream.Status == "running" {
		slog.Debug("ffmpeg already started", "id", id, "status", stream.Status, "ffmpeg", stream.FFmpeg)
		return
	}

	// Get stdout
	stdout, err := stream.cmd.StdoutPipe()
	if err != nil {
		fmt.Println("ERROR1", stream.Id, err)
	}
	go pipeStdoutToWebSocket(stdout, stream.Id)

	// // Get stderr
	// stderr, err := stream.cmd.StderrPipe()
	// if err != nil {
	// 	fmt.Println("ERROR3", err)
	// 	return
	// }
	// go pipeStderrToStdout(stderr, stream.Id)

	// Start command
	stream.Status = "running"
	slog.Info("stream started", "name", stream.Name, "status", stream.Status, "ffmpeg", stream.FFmpeg)
	err = stream.cmd.Start()
	if err != nil {
		stream.Status = "stopped"
		log.Fatal(err)
	}
}

func stopStream(id int64) {
	stream := getStreamById(id)

	// Check if the process is already terminated
    if stream.cmd.ProcessState != nil && stream.cmd.ProcessState.Exited() {
        slog.Info("stream already stopped", "name", stream.Name, "ffmpeg", stream.FFmpeg)
        return
    }

	// Stop Steam
	slog.Info("stream stopped", "name", stream.Name, "ffmpeg", stream.FFmpeg)
	stream.Status = "stopped"
	err := stream.cmd.Process.Kill()
    if err != nil {
        slog.Error("failed to kill process", "error", err)
        return
    }

	// Wait for the process to exit
	_, err = stream.cmd.Process.Wait()
	if err != nil {
		slog.Error("failed to wait for process", "error", err)
		return
	}

	// Recreate Stream
	removeStream(stream.Id)
	newStreamWithId(stream.Name, stream.FFmpeg, stream.Id)
	slog.Info("stream stopped and recreated", "name", stream.Name, "ffmpeg", stream.FFmpeg)
}

func pipeStdoutToWebSocket(stdout io.ReadCloser, id int64) {
	defer stdout.Close()

	buf := make([]byte, 65536)
	for {
		n, err := stdout.Read(buf)
		if err != nil && err != io.EOF {
			fmt.Println("ERROR2", id, err)
			return
		}

		if n > 0 {
			// fmt.Println("DATA", id, n)
			broadcastToWsClients(buf[:n], id)
		}

		if err == io.EOF {
			fmt.Println("ERROR5", id, err)
			stopStream(id)
			return
		}

	}
}

func pipeStderrToStdout(stderr io.ReadCloser, id int64) {
	defer stderr.Close()

	r := bufio.NewReader(stderr)
	for {
		line, _, err := r.ReadLine()
		if err != nil {
			if err == io.EOF {
				fmt.Println("ERROR4", id, err)
                break
            }
			fmt.Println("ERROR6", id, err)
            break
		}

		if line != nil {
			fmt.Println("ffmpeg log", id, string(line))
		}
	}
}
