package email

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"go-micro.dev/v4/config/reader"

	"gitlab.com/healthcare-integration/golang/notification-service/ent"
)

var config configuration

func Configure(r reader.Value) error {
	err := r.Scan(&config)
	if nil != err {
		return err
	}
	return nil
}

type configuration struct {
	Key     string `json:"key"`
	Name    string `json:"name"`
	ReplyTo string `json:"reply_to"`
}

type apiError struct {
	Message string `json:"message"`
	Field   string `json:"field"`
	Help    string `json:"help"`
}
type errorsList struct {
	Errors []apiError `json:"errors"`
}

func (e errorsList) Error() string {
	var array []string

	for _, v := range e.Errors {
		array = append(array, v.Message)
	}
	return strings.Join(array, ", ")
}

type Api struct {
}

//
// type Personalization struct {
//	To     []Email                `json:"to"`
//	Params map[string]interface{} `json:"dynamic_template_data"`
// }

type Attachment struct {
	Filename    string `json:"filename"`
	Content     string `json:"content"`
	Disposition string `json:"disposition"`
	Type        string `json:"type"`
}
type Body struct {
	From             *mail.Email           `json:"from"`
	Subject          string                `json:"subject"`
	ReplyTo          *mail.Email           `json:"reply_to"`
	Content          string                `json:"content"`
	TemplateId       string                `json:"template_id,omitempty"`
	Personalizations *mail.Personalization `json:"personalizations,omitempty"`
	Attachments      []interface{}         `json:"attachments,omitempty"`
}

func NewMailer() *Api {
	return new(Api)
}

func (s *Api) Do(notification *ent.Notification) (err error) {

	request := sendgrid.GetRequest(config.Key, "/v3/mail/send", "https://api.sendgrid.com")
	request.Method = "POST"

	to := mail.NewEmail(notification.Name, notification.Address.String())

	from := mail.NewEmail(notification.From, notification.From)
	replyTo := mail.NewEmail(notification.ReplyTo, notification.ReplyTo)

	body := mail.NewV3Mail()
	body.SetReplyTo(replyTo)
	body.SetFrom(from)

	if "" != notification.Body {
		content := mail.NewContent("text/html", notification.Body)
		body.AddContent(content)
	}

	personalization := mail.NewPersonalization()
	body.AddPersonalizations(personalization)
	personalization.AddTos(to)
	personalization.Subject = notification.Headline

	if nil != notification.Meta {
		body.SetTemplateID(notification.Meta.TemplateID)
		for k, v := range notification.Meta.Params {
			personalization.SetDynamicTemplateData(k, v)
		}

		if nil != notification.Meta.Attachment {
			meta := notification.Meta.Attachment
			attachment := &mail.Attachment{
				Content:     meta.Content,
				Type:        meta.Type,
				Filename:    meta.Filename,
				Disposition: meta.Disposition,
			}
			body.AddAttachment(attachment)
		}
	}

	payload := mail.GetRequestBody(body)
	request.Body = payload

	if nil != err {
		return err
	}

	response, err := sendgrid.API(request)
	if nil != err {
		return err
	}

	if response.StatusCode > 250 {
		var errData errorsList
		err = json.Unmarshal([]byte(response.Body), &errData)
		if nil != err {
			return fmt.Errorf("can't parse response body %d", response.StatusCode)
		}
		return errors.New(fmt.Sprintf("Failed with status %s", errData.Error()))
	}
	return nil
}
