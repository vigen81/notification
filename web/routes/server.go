package routes

import (
	"encoding/json"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"gitlab.com/healthcare-integration/golang/notification-service/service/broker"
	"gitlab.com/healthcare-integration/golang/notification-service/service/processing/twilio_client"
	"go-micro.dev/v4/logger"
	"net/http"

	//"github.com/twilio/twilio-go"
	"gitlab.com/healthcare-integration/common/bootstrap"
)

var engine = echo.New()

func Server() *echo.Echo {

	engine.Use(middleware.RecoverWithConfig(middleware.RecoverConfig{
		DisableStackAll: true,
	}))
	engine.Use(bootstrap.HttpWrapper())

	engine.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Response().Header().Set("Access-Control-Allow-Origin", c.Request().Header.Get("Origin"))
			c.Response().Header().Set("Access-Control-Allow-Headers", "*")
			c.Response().Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH, HEAD, OPTIONS")
			c.Response().Header().Set("Access-Control-Allow-Credentials", "true")
			c.Response().Header().Set("Allow", "GET, POST, PUT, DELETE, OPTIONS, PATCH, HEAD, OPTIONS")
			return next(c)
		}
	})

	engine.Use(middleware.Logger())
	engine.Use(middleware.CORS())
	return bind(engine)
}

type IncomingMessage struct {
	Status string `json:"status"`
	From   string `json:"from"`
	Sid    string `json:"sid"`
}
type StatusNotification struct {
	TenantID int64  `json:"tenant_id"`
	Status   string `json:"status"`
	From     string `json:"from"`
}

func bind(route *echo.Echo) *echo.Echo {

	const QueueName = "twilio_webhook_q"

	route.POST("/incoming-message", func(c echo.Context) error {
		//message := c.FormValue("Body")
		var msg IncomingMessage
		if err := c.Bind(&msg); err != nil {
			return err
		}
		if msg.Status == "stop" {
			sid := msg.Sid
			id, err := twilio_client.GetTenantBySid(sid)
			if err != nil {
				logger.Errorf("Sid not found %s", sid)
				return c.NoContent(http.StatusBadRequest)
			}
			data := StatusNotification{
				TenantID: id,
				From:     msg.From,
				Status:   msg.Status,
			}
			req, err := json.Marshal(data)
			if err != nil {
				logger.Error(err)
				return c.NoContent(http.StatusInternalServerError)
			}

			publisher, err := broker.Publisher(QueueName, "twilio-webhook_x", broker.Fanout)
			if err != nil {
				logger.Error(err)
				return c.NoContent(http.StatusInternalServerError)
			}

			err = publisher.Publish(QueueName, &message.Message{
				Payload: req,
			})
			if err != nil {
				logger.Errorf("Failed to publish status notification: %s", err.Error())
				return c.JSON(http.StatusInternalServerError, data)
			}
		}
		return c.JSON(200, "OK")
	})
	return route
}
