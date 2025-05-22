package rocketmq

import (
	"os"

	"github.com/apache/rocketmq-clients/golang/v5"
	"github.com/apache/rocketmq-clients/golang/v5/credentials"
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
	Endpoint = os.Getenv("RMQ_ENDPOINT")
	NameSpace = os.Getenv("RMQ_NAMESPACE")

	AccessKey = os.Getenv("RMQ_ACCESS_KEY")
	SecretKey = os.Getenv("RMQ_SECRET_KEY")

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
