package main

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"os"
	"strings"

	"github.com/houtarchat-cyber/cement-core"
)

//export UserLogin
func UserLogin(username, password string) error {
	return cement.UserLogin(username, password)
}

//export UserCreate
func UserCreate(username, password string) error {
	return cement.UserCreate(username, password)
}

//export GetTimestamp
func GetTimestamp() string {
	return cement.GetTimestamp()
}

//export GetCaptcha
func GetCaptcha(timestamp string) string {
	return cement.GetCaptcha(timestamp)
}

//export CheckCaptcha
func CheckCaptcha(captcha, timestamp string) bool {
	return cement.CheckCaptcha(captcha, timestamp)
}

//export Serve
func Serve(bind, keyPrefix string) error {
	return cement.Serve(bind, keyPrefix)
}

//export ChannelCreate
func ChannelCreate(channelName string) error {
	return cement.ChannelCreate(channelName)
}

//export ChannelSend
func ChannelSend(channelName, username, message string) error {
	return cement.ChannelSend(channelName, username, message)
}

//export ChannelReceive
func ChannelReceive(channelName string) ([]string, error) {
	_, jsons, err := cement.ChannelReceive(channelName)
	return jsons, err
}

//export Proxy
func Proxy(url string) error {
	return cement.Proxy(cement.GetClashConfig(url))
}

func main() {
	// 1. ./cement proxy <url>
	if len(os.Args) == 3 && os.Args[1] == "proxy" {
		err := Proxy(os.Args[2])
		if err != nil {
			panic(err)
		}
		return
	}
	// 2. ./cement serve <port>
	if len(os.Args) == 3 && os.Args[1] == "serve" {
		// login first
		err := UserLogin(os.Getenv("CMT_USER"), os.Getenv("CMT_PASS"))
		if err != nil {
			panic(err)
		}
		// serve
		err = Serve(os.Args[2], os.Getenv("CMT_USER"))
		if err != nil {
			panic(err)
		}
		return
	}
	// 3. ./cement channel create <channelName>
	if len(os.Args) == 4 && os.Args[1] == "channel" && os.Args[2] == "create" {
		err := ChannelCreate(os.Args[3])
		if err != nil {
			panic(err)
		}
		return
	}
	// 4. ./cement channel send <channelName> <message>
	if len(os.Args) == 5 && os.Args[1] == "channel" && os.Args[2] == "send" {
		err := ChannelSend(os.Args[3], os.Getenv("CMT_USER"), os.Args[4])
		if err != nil {
			panic(err)
		}
		return
	}
	// 5. ./cement channel receive <channelName>
	if len(os.Args) == 4 && os.Args[1] == "channel" && os.Args[2] == "receive" {
		jsons, _, err := cement.ChannelReceive(os.Args[3])
		if err != nil {
			panic(err)
		}
		for _, json := range jsons {
			fmt.Println(json.Username+"于"+json.Time+"发送消息：", json.Message)
		}
		return
	}
	// 6. ./cement user login <username> <password>
	if len(os.Args) == 5 && os.Args[1] == "user" && os.Args[2] == "login" {
		err := UserLogin(os.Args[3], os.Args[4])
		if err != nil {
			panic(err)
		}
		os.Setenv("CMT_USER", os.Args[3])
		os.Setenv("CMT_PASS", os.Args[4])
		return
	}
	// 7. ./cement user create <username> <password>
	if len(os.Args) == 5 && os.Args[1] == "user" && os.Args[2] == "create" {
		timestamp := GetTimestamp()
		captcha := GetCaptcha(timestamp)
		decoded, err := base64.StdEncoding.DecodeString(captcha)
		if err != nil {
			panic(err)
		}
		err = os.WriteFile("captcha.jpeg", decoded, 0644)
		if err != nil {
			panic(err)
		}

		if os.PathSeparator == '/' {
			_, err := os.Stat("/usr/bin/xdg-open")
			if err == nil {
				_, err = os.StartProcess("/usr/bin/xdg-open", []string{"xdg-open", "captcha.jpeg"}, &os.ProcAttr{})
				if err != nil {
					panic(err)
				}
			} else {
				_, err = os.Stat("/usr/bin/open")
				if err == nil {
					_, err = os.StartProcess("/usr/bin/open", []string{"open", "captcha.jpeg"}, &os.ProcAttr{})
					if err != nil {
						panic(err)
					}
				} else {
					fmt.Println("请手动打开captcha.jpeg")
				}
				if err != nil {
					panic(err)
				}
			}
		} else if os.PathSeparator == '\\' {
			_, err := os.Stat("C:\\Program Files\\Windows Photo Viewer\\PhotoViewer.dll")
			if err == nil {
				_, err = os.StartProcess("C:\\Program Files\\Windows Photo Viewer\\PhotoViewer.dll", []string{"PhotoViewer.dll", "captcha.jpeg"}, &os.ProcAttr{})
				if err != nil {
					panic(err)
				}
			} else {
				_, err = os.StartProcess("cmd", []string{"cmd", "/c", "start", "captcha.jpeg"}, &os.ProcAttr{})
				if err != nil {
					panic(err)
				}
			}
			if err != nil {
				panic(err)
			}
		}

		// check captcha
		var input string
		fmt.Print("请输入验证码：")
		fmt.Scanln(&input)
		if !CheckCaptcha(input, timestamp) {
			panic("验证码错误")
		}

		err = UserCreate(os.Args[3], os.Args[4])
		if err != nil {
			panic(err)
		}
		return
	}
	// 8. ./cement help
	if len(os.Args) == 2 && os.Args[1] == "help" {
		fmt.Println("Usage:")
		fmt.Println("  ./cement                                         进入交互模式")
		fmt.Println("  ./cement help                                    显示帮助")
		fmt.Println("  ./cement help <command>                          显示命令的帮助")
		fmt.Println("  ./cement proxy <url>                             启动代理")
		fmt.Println("  ./cement serve <port>                            启动WebDAV服务")
		fmt.Println("  ./cement channel create <channelName>            创建频道")
		fmt.Println("  ./cement channel send <channelName> <message>    发送消息")
		fmt.Println("  ./cement channel receive <channelName>           接收消息")
		fmt.Println("  ./cement user login <username> <password>        登录")
		fmt.Println("  ./cement user create <username> <password>       创建用户")
		return
	}
	// 9. ./cement help <command>
	if len(os.Args) == 3 && os.Args[1] == "help" {
		switch os.Args[2] {
		case "proxy":
			fmt.Println("Usage:")
			fmt.Println("  ./cement proxy <url>                             启动代理")
			fmt.Println("    参数url为Clash配置文件地址, 例如http://example.com/config.yaml")
			return
		case "serve":
			fmt.Println("Usage:")
			fmt.Println("  ./cement serve <port>                            启动WebDAV服务")
			fmt.Println("    参数port为端口号, 例如8080")
			return
		case "channel":
			fmt.Println("Usage:")
			fmt.Println("  ./cement channel create <channelName>            创建频道")
			fmt.Println("    参数channelName为频道名, 例如test")
			fmt.Println("  ./cement channel send <channelName> <message>    发送消息")
			fmt.Println("    参数channelName为频道名, 例如test")
			fmt.Println("    参数message为消息内容, 例如hello")
			fmt.Println("  ./cement channel receive <channelName>           接收消息")
			fmt.Println("    参数channelName为频道名, 例如test")
			return
		case "user":
			fmt.Println("Usage:")
			fmt.Println("  ./cement user login <username> <password>        登录")
			fmt.Println("    参数username为用户名, 例如test")
			fmt.Println("    参数password为密码, 例如123456")
			fmt.Println("  ./cement user create <username> <password>       创建用户")
			fmt.Println("    参数username为用户名, 例如test")
			fmt.Println("    参数password为密码, 例如123456")
			return
		default:
			fmt.Println("未知的命令")
			return
		}
	}
	if len(os.Args) == 1 {
		asciiart := " ______   ______   ___ __ __   ______   ___   __    _________  \n/_____/\\ /_____/\\ /__//_//_/\\ /_____/\\ /__/\\ /__/\\ /________/\\ \n\\:::__\\/ \\::::_\\/_\\::\\| \\| \\ \\\\::::_\\/_\\::\\_\\\\  \\ \\\\__.::.__\\/ \n \\:\\ \\  __\\:\\/___/\\\\:.      \\ \\\\:\\/___/\\\\:. `-\\  \\ \\  \\::\\ \\   \n  \\:\\ \\/_/\\\\::___\\/_\\:.\\-/\\  \\ \\\\::___\\/_\\:. _    \\ \\  \\::\\ \\  \n   \\:\\_\\ \\ \\\\:\\____/\\\\. \\  \\  \\ \\\\:\\____/\\\\. \\`-\\  \\ \\  \\::\\ \\ \n    \\_____\\/ \\_____\\/ \\__\\/ \\__\\/ \\_____\\/ \\__\\/ \\__\\/   \\__\\/ \n                                                               "
		fmt.Println(asciiart)
		for {
			fmt.Print("cement> ")
			// to read spaces, use bufio
			reader := bufio.NewReader(os.Stdin)
			input, _ := reader.ReadString('\n')
			input = strings.TrimSpace(input)
			args := strings.Split(input, " ")
			if len(args) == 0 {
				continue
			}
			os.Args = append([]string{"./cement"}, args...)
			main()
		}
	}
	fmt.Println("未知的命令")
}
