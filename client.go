package cement

import (
	"fmt"
	"os"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"gopkg.in/yaml.v2"
)

func ReadConfig() (string, string, string, string) {
	yamlFile, err := os.ReadFile("access_key.yaml")
	if err != nil {
		fmt.Println(err.Error())
	}
	var config map[string]map[string]string
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		fmt.Println(err.Error())
	}
	return config["access_key"]["endpoint"], config["access_key"]["access_key_id"], config["access_key"]["access_key_secret"], config["access_key"]["bucket_name"]
}

func GetBucket() *oss.Bucket {
	endpoint, accessKeyId, accessKeySecret, bucketName := ReadConfig()
	client, err := oss.New(endpoint, accessKeyId, accessKeySecret)
	if err != nil {
		fmt.Println("Error:", err)
	}
	bucket, err := client.Bucket(bucketName)
	if err != nil {
		fmt.Println("Error:", err)
	}
	return bucket
}
