package cement_test

import (
	"cement"
	"testing"
	"time"
)

func TestProxy(t *testing.T) {
	conf := cement.GetClashConfig("https://drive.houtar.eu.org/users/test/clash.yaml")
	go cement.Proxy(conf)
	time.Sleep(500 * time.Millisecond)
}
