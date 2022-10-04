package cement

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"io/ioutil"
	"strconv"
	"strings"
	"time"
)

func ChannelCreate(channelname string) error {
	bucket := GetBucket()
	channelname = base64.StdEncoding.EncodeToString([]byte(channelname))

	isExist, err := bucket.IsObjectExist("channels/" + channelname)
	if err != nil {
		return err
	}
	if isExist {
		return errors.New("channel already exist")
	}

	_, err = bucket.AppendObject("channels/"+channelname, strings.NewReader(""), 0)

	return err
}

func ChannelSend(channelname string, username string, message string) error {
	bucket := GetBucket()
	channelname = base64.StdEncoding.EncodeToString([]byte(channelname))
	username = base64.StdEncoding.EncodeToString([]byte(username))
	message = base64.StdEncoding.EncodeToString([]byte(message))

	isExist, err := bucket.IsObjectExist("channels/" + channelname)
	if err != nil {
		return err
	}
	if !isExist {
		return errors.New("channel not exist")
	}

	json := "{\"time\":\"" + time.Now().Format("2006-01-02 15:04:05") + "\",\"username\":\"" + username + "\",\"message\":\"" + message + "\"}\n"

	props, err := bucket.GetObjectDetailedMeta("channels/" + channelname)
	if err != nil {
		return err
	}
	nextPos, err := strconv.ParseInt(props.Get("X-Oss-Next-Append-Position"), 10, 64)
	if err != nil {
		return err
	}
	_, err = bucket.AppendObject("channels/"+channelname, strings.NewReader(json), nextPos)

	return err
}

type Messages struct {
	Time     string `json:"time"`
	Username string `json:"username"`
	Message  string `json:"message"`
}

func ChannelReceive(channelname string) ([]Messages, error) {
	bucket := GetBucket()
	channelname = base64.StdEncoding.EncodeToString([]byte(channelname))

	isExist, err := bucket.IsObjectExist("channels/" + channelname)
	if err != nil {
		return nil, err
	}
	if !isExist {
		return nil, errors.New("channel not exist")
	}

	channelfile, err := bucket.GetObject("channels/" + channelname)
	if err != nil {
		return nil, err
	}

	defer channelfile.Close()

	channel, err := ioutil.ReadAll(channelfile)
	if err != nil {
		return nil, err
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

	return messageList, nil
}
