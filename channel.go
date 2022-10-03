package cement

import (
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"strconv"
	"strings"
	"time"
)

func channelCreate(channelname string) string {
	bucket := getBucket()
	channelname = base64.StdEncoding.EncodeToString([]byte(channelname))

	isExist, err := bucket.IsObjectExist("channels/" + channelname)
	if err != nil {
		return err.Error()
	}
	if isExist {
		return "channel already exist"
	}

	_, err = bucket.AppendObject("channels/"+channelname, strings.NewReader(""), 0)

	return "create success"
}

func channelSend(channelname string, username string, message string) string {
	bucket := getBucket()
	channelname = base64.StdEncoding.EncodeToString([]byte(channelname))
	username = base64.StdEncoding.EncodeToString([]byte(username))
	message = base64.StdEncoding.EncodeToString([]byte(message))

	isExist, err := bucket.IsObjectExist("channels/" + channelname)
	if err != nil {
		return err.Error()
	}
	if !isExist {
		return "channel not exist"
	}

	json := "{\"time\":\"" + time.Now().Format("2006-01-02 15:04:05") + "\",\"username\":\"" + username + "\",\"message\":\"" + message + "\"}\n"

	props, err := bucket.GetObjectDetailedMeta("channels/" + channelname)
	if err != nil {
		return err.Error()
	}
	nextPos, err := strconv.ParseInt(props.Get("X-Oss-Next-Append-Position"), 10, 64)
	if err != nil {
		return err.Error()
	}
	_, err = bucket.AppendObject("channels/"+channelname, strings.NewReader(json), nextPos)

	return "send success"
}

type Messages struct {
	Time     string `json:"time"`
	Username string `json:"username"`
	Message  string `json:"message"`
}

func channelReceive(channelname string) ([]Messages, string) {
	bucket := getBucket()
	channelname = base64.StdEncoding.EncodeToString([]byte(channelname))

	isExist, err := bucket.IsObjectExist("channels/" + channelname)
	if err != nil {
		return nil, err.Error()
	}
	if !isExist {
		return nil, "channel not exist"
	}

	channelfile, err := bucket.GetObject("channels/" + channelname)
	if err != nil {
		return nil, err.Error()
	}

	defer channelfile.Close()

	channel, err := ioutil.ReadAll(channelfile)
	if err != nil {
		return nil, err.Error()
	}

	var messageList []Messages

	for _, line := range strings.Split(string(channel), "\n") {
		var message Messages
		err = json.Unmarshal([]byte(line), &message)
		if err != nil {
			break
		}
		username, _ := base64.StdEncoding.DecodeString(message.Username)
		messageText, _ := base64.StdEncoding.DecodeString(message.Message)
		message.Username = string(username)
		message.Message = string(messageText)
		messageList = append(messageList, message)
	}

	return messageList, "receive success"
}
