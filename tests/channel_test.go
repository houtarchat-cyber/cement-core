package cement_test

import (
	"testing"

	"github.com/houtarchat-cyber/cement-core"
)

func TestChannel(t *testing.T) {
	timestamp := cement.GetTimestamp()
	// create a cement channel
	err := cement.ChannelCreate(timestamp)
	if err != nil {
		t.Error(err)
	}
	// send a message to cement channel
	err = cement.ChannelSend(timestamp, "peter", "hello")
	if err != nil {
		t.Error(err)
	}
	// get messages from cement channel
	messages, _, err := cement.ChannelReceive(timestamp)
	if err != nil {
		t.Error(err)
	}
	if len(messages) != 1 {
		t.Error("messages count error")
	}
	if messages[0].Username != "peter" {
		t.Error("username error")
	}
	if messages[0].Message != "hello" {
		t.Error("message error")
	}
}
