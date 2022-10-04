package cement_test

import (
	"cement"
	"testing"
)

func TestClient(t *testing.T) {
	// read config
	endpoint, accessKeyId, accessKeySecret, bucketName := cement.ReadConfig()
	if endpoint == "" || accessKeyId == "" || accessKeySecret == "" || bucketName == "" {
		t.Error("read config error")
	}
	// get bucket
	bucket := cement.GetBucket()
	if bucket == nil {
		t.Error("get bucket error")
	}
}
