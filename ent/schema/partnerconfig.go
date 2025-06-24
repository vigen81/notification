package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"entgo.io/ent/schema/mixin"
)

// PartnerConfig holds the schema definition for the PartnerConfig entity.
type PartnerConfig struct {
	ent.Schema
}

// Fields of the PartnerConfig.
func (PartnerConfig) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").Default(""),
		field.Int64("tenant_id").Unique(),
		field.JSON("email_providers", []byte{}),
		field.JSON("sms_providers", []byte{}),
		field.JSON("push_providers", []byte{}),
		field.JSON("batch_config", []byte{}),
		field.JSON("rate_limits", []byte{}),
		field.Bool("enabled").Default(true),
	}
}

func (PartnerConfig) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Time{},
	}
}

// Edges of the PartnerConfig.
func (PartnerConfig) Edges() []ent.Edge {
	return nil
}

func (PartnerConfig) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id"),
		index.Fields("enabled"),
	}
}
