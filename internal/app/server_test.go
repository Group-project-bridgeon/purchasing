package app

import (
	"context"
	"io"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
)

func TestListenerFailureReturnsInsteadOfWaitingForSignal(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	start := time.Now()
	err = Serve(context.Background(), fiber.New(), listener.Addr().String(), time.Second)
	if err == nil || time.Since(start) > time.Second {
		t.Fatal("listen failure did not terminate promptly")
	}
}

func TestShutdownDrainsAcceptedHTTPBeforeReturning(t *testing.T) {
	server := fiber.New(fiber.Config{DisableStartupMessage: true, ReadTimeout: time.Second, WriteTimeout: 2 * time.Second, IdleTimeout: time.Second})
	address := make(chan string, 1)
	server.Hooks().OnListen(func(data fiber.ListenData) error { address <- net.JoinHostPort(data.Host, data.Port); return nil })
	entered, release := make(chan struct{}), make(chan struct{})
	server.Get("/work", func(c *fiber.Ctx) error { close(entered); <-release; return c.SendString("completed") })
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	result := make(chan error, 1)
	go func() { result <- Serve(ctx, server, "127.0.0.1:0", 2*time.Second) }()
	var url string
	select {
	case a := <-address:
		url = "http://" + a + "/work"
	case <-time.After(3 * time.Second):
		t.Fatal("owned listener did not start")
	}
	response := make(chan error, 1)
	go func() {
		client := http.Client{Timeout: 3 * time.Second}
		resp, err := client.Get(url)
		if err == nil {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			if resp.StatusCode != 200 || string(body) != "completed" {
				err = context.Canceled
			}
		}
		response <- err
	}()
	select {
	case <-entered:
	case <-time.After(time.Second):
		t.Fatal("request was not accepted")
	}
	cancel()
	select {
	case <-result:
		t.Fatal("shutdown returned before accepted HTTP completed")
	case <-time.After(20 * time.Millisecond):
	}
	close(release)
	select {
	case err := <-response:
		if err != nil {
			t.Fatal("accepted request did not finish")
		}
	case <-time.After(3 * time.Second):
		t.Fatal("request drain exceeded bound")
	}
	select {
	case err := <-result:
		if err != nil {
			t.Fatal("graceful shutdown failed")
		}
	case <-time.After(3 * time.Second):
		t.Fatal("shutdown exceeded bound")
	}
}
