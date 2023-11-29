package twilio_client

import (
	`errors`
	`sync`
	
	"github.com/twilio/twilio-go"
	openapi `github.com/twilio/twilio-go/rest/api/v2010`
	`go-micro.dev/v4/config/reader`
)

var params Config

type ISms interface {
	Sms(to, body, service string) (string, error)
}

func Configure(r reader.Value) error {
	err := r.Scan(&params)
	if nil != err {
		return err
	}
	getClient()
	return nil
}

type Config struct {
	AccountSid string            `json:"account_sid"`
	AuthToken  string            `json:"auth_token"`
	From       map[string]string `json:"from"`
}

func Sms(to, payload, service string) (data string, err error) {
	from, ok := params.From[service]
	if false == ok {
		return "", errors.New("service not found")
	}
	body := &openapi.CreateMessageParams{}
	body.SetTo(to)
	body.SetMessagingServiceSid(from)
	body.SetBody(payload)
	resp, err := getClient().Client.ApiV2010.CreateMessage(body)
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

func getClient() *httpClient {
	once.Do(func() {
		client := twilio.NewRestClientWithParams(twilio.RestClientParams{
			Username: params.AccountSid,
			Password: params.AuthToken,
		})
		api = &httpClient{Client: client}
	})
	return api
}
