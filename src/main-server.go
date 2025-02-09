package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gorilla/websocket"
)

type WsClient struct {
	conn     *websocket.Conn
	streamId int64
}

var wsClients = make([]*WsClient, 0)

var wsUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

// HTTP Server

func httpServerStart() error {
	mux := http.NewServeMux()

	// Routes
	mux.Handle("GET /api/streams", middlewareLogger(http.HandlerFunc(handlerGetStreams)))
	mux.Handle("/api/stream/start/{id}", middlewareLogger(http.HandlerFunc(handlerStartStream)))
	mux.Handle("/api/stream/stop/{id}", middlewareLogger(http.HandlerFunc(handlerStopStream)))
	mux.Handle("/api/stream/remove/{id}", middlewareLogger(http.HandlerFunc(handlerRemoveStream)))
	mux.Handle("DELETE /api/streams", middlewareLogger(http.HandlerFunc(handlerClearStreams)))
	mux.Handle("POST /api/stream/new", middlewareLogger(http.HandlerFunc(handlerNewStream)))
	mux.Handle("POST /api/stream/new/start", middlewareLogger(http.HandlerFunc(handlerNewStreamStart)))
	mux.Handle("POST /api/stream/update/name/{id}", middlewareLogger(http.HandlerFunc(handlerUpdateStreamName)))
	mux.Handle("/add", middlewareLogger(http.HandlerFunc(handlerAdd)))
	mux.Handle("/view", middlewareLogger(http.HandlerFunc(handlerView)))
	mux.Handle("/ws/{id}", middlewareLogger(http.HandlerFunc(handlerWs)))

	// Public Folder Routes
	mux.Handle("/public/", middlewareLogger(http.StripPrefix("/public/", http.FileServer(http.Dir("../public")))))
	mux.Handle("/", middlewareLogger(http.HandlerFunc(handlerBaseUrl)))

	// Get HTTP Port
	port := os.Getenv("HTTP_PORT")
	if port == "" {
		port = "9000"
	}
	port = fmt.Sprintf(":%s", port)

	// Start Web Server
	slog.Info("http server started", "url", "http://localhost"+port)
	err := http.ListenAndServe(port, mux)

	return err
}

func handlerBaseUrl(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/view?x=1&y=1", http.StatusSeeOther)
}

func handlerGetStreams(w http.ResponseWriter, r *http.Request) {
	streams := getStreams()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(streams)
}

func handlerStartStream(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		http.NotFound(w, r)
		return
	}
	startStream(int64(id))
	w.Write([]byte("ok"))
}

func handlerStopStream(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		http.NotFound(w, r)
		return
	}
	stopStream(int64(id))
	w.Write([]byte("ok"))
}

func handlerRemoveStream(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		http.NotFound(w, r)
		return
	}
	removeStream(int64(id))
	w.Header().Add("HX-Refresh", "true")
	w.Write([]byte("ok"))
}

func handlerClearStreams(w http.ResponseWriter, r *http.Request) {
	clearAllStreams()
	w.Write([]byte("ok"))
}

func handlerNewStream(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") == "application/json" {

		// JSON encoded
		var s Stream
		err := json.NewDecoder(r.Body).Decode(&s)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		newStream(s.Name, s.FFmpeg)
		w.Write([]byte("ok"))

	} else {

		// Form (url-encoded)
		name := r.Form.Get("name")
		r.Form.Del("name")
		ffmpeg := r.Form.Get("ffmpeg")
		r.Form.Del("ffmpeg")
		newStream(name, ffmpeg)
		w.Write([]byte("ok"))

	}
}

func handlerNewStreamStart(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") == "application/json" {

		// JSON encoded
		var s Stream
		err := json.NewDecoder(r.Body).Decode(&s)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		stream := newStream(s.Name, s.FFmpeg)
		startStream(stream.Id)
		http.Redirect(w, r, "/view?x=1&y=1", http.StatusSeeOther)
		w.Write([]byte("ok"))

	} else {

		// Form (url-encoded)
		name := r.Form.Get("name")
		r.Form.Del("name")
		ffmpeg := r.Form.Get("ffmpeg")
		ffmpeg = strings.ReplaceAll(ffmpeg, "\r\n", " ")
		ffmpeg = strings.ReplaceAll(ffmpeg, "\n", " ")
		r.Form.Del("ffmpeg")
		stream := newStream(name, ffmpeg)
		startStream(stream.Id)
		http.Redirect(w, r, "/view?x=1&y=1", http.StatusSeeOther)
		w.Write([]byte("ok"))

	}
}

func handlerUpdateStreamName(w http.ResponseWriter, r *http.Request) {

	// Get Id From Path
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		http.NotFound(w, r)
		return
	}
	id64 := int64(id)

	if r.Header.Get("Content-Type") == "application/json" {

		// JSON encoded
		var s Stream
		err := json.NewDecoder(r.Body).Decode(&s)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		updateStreamName(id64, s.Name)
		w.Write([]byte("ok"))

	} else {

		// Form (url-encoded)
		name := r.Form.Get("name")
		r.Form.Del("name")
		updateStreamName(id64, name)
		w.Write([]byte("ok"))

	}
}

// HTMX Templates

func parse_templates(layout string, page string) *template.Template {
	tmp, err := template.New("").ParseGlob("./templates/comp*")
	if err != nil {
		log.Fatal(err)
	}

	tmp, err = tmp.ParseFiles("./templates/" + layout + ".html")
	if err != nil {
		log.Fatal(err)
	}

	tmp, err = tmp.ParseFiles("./templates/" + page + ".html")
	if err != nil {
		log.Fatal(err)
	}

	return tmp
}

func handlerView(w http.ResponseWriter, r *http.Request) {
	tmp := parse_templates("layout-header", "page-view")

	x, err := strconv.Atoi(r.URL.Query().Get("x"))
	if err != nil {
		x = 1
	}
	if x > 4 {
		x = 4
	}

	y, err := strconv.Atoi(r.URL.Query().Get("y"))
	if err != nil {
		y = 1
	}
	if y > 4 {
		y = 4
	}

	streams := getStreams()

	gridX := x
	gridY := y

	diff := (gridX * gridY) - len(streams)

	watching := streams

	if diff > 0 {
		for i := 1; i <= diff; i++ {
			watching = append(watching, &Stream{Id: 0})
		}
	} else {
		watching = watching[:len(watching)+diff]
	}

	tmp.ExecuteTemplate(w, "layout-header", struct {
		Title    string
		Streams  []*Stream
		Watching []*Stream
		GridX    int
		GridY    int
	}{
		Title:    "Title",
		Streams:  streams,
		Watching: watching,
		GridX:    gridX,
		GridY:    gridY,
	})
}

func handlerAdd(w http.ResponseWriter, r *http.Request) {
	tmp := parse_templates("layout-header", "page-add")
	tmp.ExecuteTemplate(w, "layout-header", struct {
		Title string
	}{
		Title: "Title",
	})
}

// WebSocket Clients

func handlerWs(w http.ResponseWriter, r *http.Request) {

	// Get stream id from path
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		http.NotFound(w, r)
		return
	}
	id64 := int64(id)
	slog.Debug("websocket ask", "id", id)

	// Upgrade the HTTP connection to a WebSocket connection
	conn, err := wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("websocket upgrade error", "error", err)
	}
	slog.Debug("websocket client connected")

	for _, stream := range getStreams() {
		if stream.Id == id64 {
			addWsClient(conn, id64)
			go wsListener(conn)
			return
		}
	}

	slog.Debug("websocket client removed:", "error", "bad id")
	conn.Close()
}

func wsListener(conn *websocket.Conn) {
	defer conn.Close()
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			slog.Debug("websocket client disconnected:", "error", err)
			removeWsClient(conn)
			return
		}
	}
}

func addWsClient(conn *websocket.Conn, streamId int64) {
	client := WsClient{
		conn:     conn,
		streamId: streamId,
	}
	slog.Info("websocket client added", "wsClients count", len(wsClients), "streamId", streamId)
	wsClients = append(wsClients, &client)
}

func removeWsClient(conn *websocket.Conn) {
	for i, client := range wsClients {
		if client.conn == conn {
			wsClients = append(wsClients[:i], wsClients[i+1:]...)
			slog.Info("websocket client removed", "wsClients count", len(wsClients), "streamId", client.streamId)
			return
		}
	}
}

func broadcastToWsClients(buf []byte, streamId int64) {
	for _, client := range wsClients {
		if client.streamId == streamId {
			err := client.conn.WriteMessage(websocket.BinaryMessage, buf)
			if err != nil {
				slog.Error("websocket broadcast error", "error", err)
				removeWsClient(client.conn)
			}
		}
	}
}

func middlewareLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/favicon.ico" {
			r.ParseForm()
			slog.Debug("http request",
				"method", r.Method,
				"url", r.URL.Path,
				"form", r.Form,
				// "query", r.URL.RawQuery,
				// "ip", r.RemoteAddr,
			)
		}
		next.ServeHTTP(w, r)
	})
}
