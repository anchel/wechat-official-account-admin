package rocketmq

import (
	"os"

	"github.com/apache/rocketmq-clients/golang/v5"
	"github.com/apache/rocketmq-clients/golang/v5/credentials"
	"github.com/spf13/viper"
)

var (
	Endpoint  = ""
	NameSpace = ""
	Topic     = ""
	GroupName = ""

	AccessKey = ""
	SecretKey = ""
)

func InitRocketMQ() {
	Endpoint = viper.GetString("RMQ_ENDPOINT")
	NameSpace = viper.GetString("RMQ_NAMESPACE")

	AccessKey = viper.GetString("RMQ_ACCESS_KEY")
	SecretKey = viper.GetString("RMQ_SECRET_KEY")

	// log.Info("rocketmq config", "endpoint", Endpoint, "namespace", NameSpace, "topic", Topic, "groupName", GroupName, "accesskey", AccessKey, "secretkey", SecretKey)

	// log to console
	os.Setenv("rocketmq.client.logLevel", "ERROR")
	os.Setenv("mq.consoleAppender.enabled", "true")
	golang.ResetLogger()
}

func NewProducer(topic string) (golang.Producer, error) {
	// new producer instance
	producer, err := golang.NewProducer(&golang.Config{
		Endpoint:  Endpoint,
		NameSpace: NameSpace,

		Credentials: &credentials.SessionCredentials{
			AccessKey:    AccessKey,
			AccessSecret: SecretKey,
		},
	},
		golang.WithTopics(topic),
	)
	if err != nil {
		return nil, err
	}
	return producer, nil
}
