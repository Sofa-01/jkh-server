package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
    "entgo.io/ent/schema/edge"
)

// Role holds the schema definition for the Role entity.
type Role struct {
	ent.Schema
}

// Fields of the Role.
func (Role) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").Unique(),
	}
}

// Edges of the Role.
func (Role) Edges() []ent.Edge {
	return []ent.Edge{
		// Обратная связь "один ко многим": одна роль может быть у многих пользователей
		edge.To("users", User.Type),
	}
}
