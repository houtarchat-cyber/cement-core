package cement_test

import (
	"cement"
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
