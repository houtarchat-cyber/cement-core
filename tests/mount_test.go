package cement_test

import (
	"github.com/houtarchat-cyber/cement-core"
	"testing"
	"time"
)

func TestMount(t *testing.T) {
	go func() {
		err := cement.Serve(":8080", "test")
		if err != nil {
			t.Error(err)
		}
	}()
	time.Sleep(500 * time.Millisecond)
}
