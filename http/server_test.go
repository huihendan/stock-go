package http

import (
	"testing"
	"time"
)

func TestHttpServer(t *testing.T) {
	startServer()

	time.After(30 * time.Minute)
}
