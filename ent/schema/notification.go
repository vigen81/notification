package schema

import (
	"encoding/json"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"entgo.io/ent/schema/mixin"

	"gitlab.smartbet.am/golang/notification/types"
)

// Notification holds the schema definition for the Notification entity.
type Notification struct {
	ent.Schema
}

type Attachment struct {
	Filename    string `json:"filename"`
	Content     string `json:"content"`
	Disposition string `json:"disposition"`
	Type        string `json:"type"`
}

type NotificationMeta struct {
	Service    string                 `json:"service,omitempty"`
	TemplateID string                 `json:"template_id,omitempty"`
	Params     map[string]interface{} `json:"params,omitempty"`
	Attachment *Attachment            `json:"attachment,omitempty"`
	Data       json.RawMessage        `json:"data,omitempty"`
}

// Fields of the Notification.
func (Notification) Fields() []ent.Field {
	return []ent.Field{
		field.Text("body"),
		field.Int64("tenant_id"),
		field.String("headline").Optional(),
		field.String("name").Optional(),
		field.String("from").Optional(),
		field.String("reply_to").Optional(),
		field.String("tag").Optional(),
		field.Text("address").GoType(types.Address("")),
		field.String("request_id").Unique(),
		field.Int64("schedule_ts").Optional().Nillable(),
		field.Enum("type").Values("SMS", "EMAIL", "PUSH"),
		field.Enum("status").Values("ACTIVE", "COMPLETED", "CANCEL", "PENDING", "FAILED").Default("PENDING"),
		field.JSON("meta", &NotificationMeta{}).Optional(),
		field.String("error_message").Nillable().Optional(),
		field.String("batch_id").Optional(),
		field.Int("retry_count").Default(0),
	}
}

func (Notification) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Time{},
	}
}

// Edges of the Notification.
func (Notification) Edges() []ent.Edge {
	return nil
}

func (Notification) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("request_id").Annotations(entsql.Prefix(20)),
		index.Fields("schedule_ts", "status"),
		index.Fields("tenant_id", "status"),
		index.Fields("batch_id"),
		index.Fields("type", "status"),
	}
}
