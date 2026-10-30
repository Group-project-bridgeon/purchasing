package app

import (
	"context"
	"errors"
	"net"
	"time"

	"github.com/gofiber/fiber/v2"
)

// Serve binds synchronously, so listen failures reach main instead of leaving a
// process waiting for a shutdown signal. HTTP drains before dependencies close.
func Serve(ctx context.Context, server *fiber.App, address string, timeout time.Duration) error {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}
	defer listener.Close()
	result := make(chan error, 1)
	go func() { result <- server.Listener(listener) }()
	select {
	case err := <-result:
		if err == nil {
			return errors.New("HTTP listener exited unexpectedly")
		}
		return err
	case <-ctx.Done():
		shutdown, stop := context.WithTimeout(context.Background(), timeout)
		defer stop()
		if err := server.ShutdownWithContext(shutdown); err != nil {
			return err
		}
		select {
		case <-result:
			return nil
		case <-shutdown.Done():
			return shutdown.Err()
		}
	}
}
