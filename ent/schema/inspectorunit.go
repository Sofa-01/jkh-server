package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
    "entgo.io/ent/schema/edge"
)

// InspectorUnit holds the schema definition for the InspectorUnit entity.
type InspectorUnit struct {
	ent.Schema
}

// Fields of the InspectorUnit.
func (InspectorUnit) Fields() []ent.Field {
	return []ent.Field{
		// Явное определение ФК
		field.Int("user_id"), 
		field.Int("jkh_unit_id"),
	}
}

// Edges of the InspectorUnit.
func (InspectorUnit) Edges() []ent.Edge {
	return []ent.Edge{
		// Связь М:1 к Пользователю
		edge.From("inspector", User.Type).
			Ref("assigned_units"). // Обратная ссылка в User
			Unique().
			Required().
			Field("user_id"),

		// Связь М:1 к ЖЭУ
		edge.From("jkh_unit", JkhUnit.Type).
			Ref("assigned_inspectors"). // Обратная ссылка в JkhUnit
			Unique().
			Required().
			Field("jkh_unit_id"),
	}
}
