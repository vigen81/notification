package twilio_client

import (
	"errors"
	"sync"

	"github.com/twilio/twilio-go"
	openapi "github.com/twilio/twilio-go/rest/api/v2010"
	"go-micro.dev/v4/config/reader"
)

var Configs Config

type ISms interface {
	Sms(to, body, from string) (string, error)
}

func Configure(r reader.Value) error {
	var configData ConfigData
	err := r.Scan(&configData)
	Configs = Config{}
	for _, c := range configData {
		Configs[c.TenantID] = c
	}
	if nil != err {
		return err
	}
	return nil
}

type Config map[int64]ConfigItem
type ConfigData []ConfigItem
type ConfigItem struct {
	AccountSid string `json:"account_sid"`
	AuthToken  string `json:"auth_token"`
	Sid        string `json:"sid"`
	TenantID   int64  `json:"tenant_id"`
}

func GetTenantBySid(sid string) (int64, error) {
	for tenantID, config := range Configs {
		if config.Sid == sid {
			return tenantID, nil
		}
	}
	return 0, errors.New("config not found")
}

func Sms(tenantId int64, to, payload, service string) (data string, err error) {

	body := &openapi.CreateMessageParams{}
	body.SetTo(to)
	body.SetMessagingServiceSid(service)
	body.SetBody(payload)

	params, ok := Configs[tenantId]
	if !ok {
		return "", errors.New("no config found")
	}
	client := twilio.NewRestClientWithParams(twilio.ClientParams{
		Username: params.AccountSid,
		Password: params.AuthToken,
	})

	resp, err := client.Api.CreateMessage(body)
	if nil != err {
		return "", err
	}
	return *resp.Sid, nil
}

type httpClient struct {
	Client *twilio.RestClient
}

var once sync.Once

var api *httpClient
