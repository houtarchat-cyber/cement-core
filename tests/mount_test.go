package cement_test

import (
	"cement"
	"testing"
	"time"
)

func TestMount(t *testing.T) {
	go cement.Serve(":8080", "test")
	time.Sleep(500 * time.Millisecond)
}
