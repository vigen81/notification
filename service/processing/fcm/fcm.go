package fcm

import (
	`encoding/json`
	`sync`
	
	"github.com/appleboy/go-fcm"
	`go-micro.dev/v4/config/reader`
	`go-micro.dev/v4/logger`
)

var apiClient *fcm.Client

type config struct {
	Key string `json:"key"`
}

var params config

func Configure(r reader.Value) error {
	err := r.Scan(&params)
	if nil != err {
		return err
	}
	getClient()
	return nil
}

var once sync.Once

func getClient() *fcm.Client {
	once.Do(func() {
		var err error
		apiClient, err = fcm.NewClient(params.Key)
		if nil != err {
			logger.Error("wrong api")
		}
	})
	return apiClient
}

type Message struct {
	Address string
	Title   string
	Body    string
	Data    json.RawMessage `json:"data"`
}

func Send(message Message) error {
	msg := &fcm.Message{
		To:               message.Address,
		Priority:         "high",
		ContentAvailable: true,
		Notification: &fcm.Notification{
			Title: message.Title,
			Body:  message.Body,
			Sound: "default",
		},
		Data: map[string]interface{}{
			"priority":          "high",
			"content_available": true,
			"data":              message.Data,
		},
	}
	
	resp, err := getClient().Send(msg)
	
	if nil != err {
		return err
	}
	if 0 == resp.Success {
		logger.Debug(resp.Results)
		return resp.Results[0].Error
	}
	return err
}
