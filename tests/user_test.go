package cement_test

import (
	"github.com/houtarchat-cyber/cement-core"
	"testing"
)

func TestUser(t *testing.T) {
	timestamp := cement.GetTimestamp()

	// login with a non-exist user
	err := cement.UserLogin(timestamp, "test")
	if err == nil {
		t.Error("login with a non-exist user")
	}
	// create a user
	err = cement.UserCreate(timestamp, "test")
	if err != nil {
		t.Error(err)
	}
	// create a user again
	err = cement.UserCreate(timestamp, "test")
	if err == nil {
		t.Error("create a user again")
	}
	// login with a wrong password
	err = cement.UserLogin(timestamp, "test1")
	if err == nil {
		t.Error("login with a wrong password")
	}
	// login with an exist user
	err = cement.UserLogin(timestamp, "test")
	if err != nil {
		t.Error(err)
	}
}
