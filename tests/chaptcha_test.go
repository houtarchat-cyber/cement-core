package cement_test

import (
	"github.com/houtarchat-cyber/cement-core"
	"testing"
)

func TestCaptcha(t *testing.T) {
	timestamp := cement.GetTimestamp()
	captcha := cement.GetCaptcha(timestamp)
	if captcha == "" {
		t.Error("captcha is empty")
	}
	if cement.CheckCaptcha(captcha, timestamp) {
		t.Error("captcha check failed")
	}
}
