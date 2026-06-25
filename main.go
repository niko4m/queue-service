package main

import (
	"context"
	"flag"
	"net/http"
	"strings"
	"sync"
	"time"
)

const (
	valueName      = "v"
	timeoutName    = "timeout"
	defaultPort    = "8080"
	defaultTimeout = time.Second * 30
)

var mu sync.Mutex

// queue очередь сообщений, где ключ - имя очереди, а значение - канал для передачи сообщений.
var queue = make(map[string]chan string)

func server(port string) error {
	http.HandleFunc("/", Queue)
	return http.ListenAndServe(":"+port, nil)
}

func main() {
	port := flag.String("port", defaultPort, "Port to listen on")
	flag.Parse()
	server(*port)
}

func getQueueName(r *http.Request) string {
	path := r.URL.Path
	paths := strings.Split(path, "/")
	return strings.TrimSpace(paths[1])
}

func Queue(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPut:
		put(w, r)
	case http.MethodGet:
		get(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func put(w http.ResponseWriter, r *http.Request) {
	queueName := getQueueName(r)
	value := strings.TrimSpace(r.URL.Query().Get(valueName))
	if queueName == "" || value == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	chQueue := getQueue(queueName)

	go func() {
		chQueue <- value
	}()

	w.WriteHeader(http.StatusOK)
}

func get(w http.ResponseWriter, r *http.Request) {
	queueName := getQueueName(r)
	if queueName == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	timeoutDuration := defaultTimeout
	if timeout := strings.TrimSpace(r.URL.Query().Get(timeoutName)); timeout != "" {
		timeoutQueryDur, err := time.ParseDuration(timeout)
		if err == nil {
			timeoutDuration = timeoutQueryDur
		}
	}

	chQueue := getQueue(queueName)

	ctx, cancel := context.WithTimeout(r.Context(), timeoutDuration)
	defer cancel()

	select {
	case value := <-chQueue:
		w.Write([]byte(value))
	case <-ctx.Done():
		w.WriteHeader(http.StatusNotFound)
	}
}

func getQueue(queueName string) chan string {
	mu.Lock()
	defer mu.Unlock()
	chQueue, ok := queue[queueName]
	if !ok {
		chQueue = make(chan string)
		queue[queueName] = chQueue
	}
	return chQueue
}
