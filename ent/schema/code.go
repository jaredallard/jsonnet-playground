package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// Code holds the schema definition for the Code entity.
type Code struct {
	ent.Schema
}

// Fields of the Code.
func (Code) Fields() []ent.Field {
	return []ent.Field{
		// id is the ID of this code snippet
		field.UUID("id", uuid.UUID{}).Default(uuid.New).Unique().Immutable(),

		// contents is the content of this code snippet
		// For now there's a limit of 400,000 characters
		field.String("contents").NotEmpty().Immutable().MaxLen(400000),
	}
}

// Edges of the Code.
func (Code) Edges() []ent.Edge {
	return nil
}
