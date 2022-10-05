package cement

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"strconv"
	"strings"
	"time"
)

func ChannelCreate(channelName string) error {
	bucket := getBucket()
	channelName = base64.StdEncoding.EncodeToString([]byte(channelName))

	isExist, err := bucket.IsObjectExist("channels/" + channelName)
	if err != nil {
		return err
	}
	if isExist {
		return errors.New("channel already exist")
	}

	_, err = bucket.AppendObject("channels/"+channelName, strings.NewReader(""), 0)

	return err
}

func ChannelSend(channelName string, username string, message string) error {
	bucket := getBucket()
	channelName = base64.StdEncoding.EncodeToString([]byte(channelName))
	username = base64.StdEncoding.EncodeToString([]byte(username))
	message = base64.StdEncoding.EncodeToString([]byte(message))

	isExist, err := bucket.IsObjectExist("channels/" + channelName)
	if err != nil {
		return err
	}
	if !isExist {
		return errors.New("channel not exist")
	}

	j := "{\"time\":\"" + time.Now().Format("2006-01-02 15:04:05") + "\",\"username\":\"" + username + "\",\"message\":\"" + message + "\"}\n"

	props, err := bucket.GetObjectDetailedMeta("channels/" + channelName)
	if err != nil {
		return err
	}
	nextPos, err := strconv.ParseInt(props.Get("X-Oss-Next-Append-Position"), 10, 64)
	if err != nil {
		return err
	}
	_, err = bucket.AppendObject("channels/"+channelName, strings.NewReader(j), nextPos)

	return err
}

type Messages struct {
	Time     string `json:"time"`
	Username string `json:"username"`
	Message  string `json:"message"`
}

func ChannelReceive(channelName string) ([]Messages, []string, error) {
	bucket := getBucket()
	channelName = base64.StdEncoding.EncodeToString([]byte(channelName))

	isExist, err := bucket.IsObjectExist("channels/" + channelName)
	if err != nil {
		return nil, nil, err
	}
	if !isExist {
		return nil, nil, errors.New("channel not exist")
	}

	channelFile, err := bucket.GetObject("channels/" + channelName)
	if err != nil {
		return nil, nil, err
	}

	defer func(channelFile io.ReadCloser, err *error) {
		*err = channelFile.Close()
	}(channelFile, &err)
	if err != nil {
		return nil, nil, err
	}

	channel, err := io.ReadAll(channelFile)
	if err != nil {
		return nil, nil, err
	}

	var messageList []Messages
	jsons := strings.Split(string(channel), "\n")

	for _, line := range jsons {
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

	return messageList, jsons, nil
}
