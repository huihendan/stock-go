package http

import (
	"testing"
	"time"
)

func TestHttpServer(t *testing.T) {
	StartServer()

	time.After(30 * time.Minute)
}
