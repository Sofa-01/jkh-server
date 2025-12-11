package schema

import (
    "entgo.io/ent"
    "entgo.io/ent/schema/field"
    "entgo.io/ent/schema/edge"
)

// District holds the schema definition for the District entity.
type District struct {
	ent.Schema
}

// Fields of the District.
func (District) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			Unique().
			Comment("Название района (уникальное)."),
	}
}

// Edges of the District.
func (District) Edges() []ent.Edge {
	return []ent.Edge{
		// Обратная связь 1:М к ЖЭУ (JkhUnits)
		edge.To("jkh_units", JkhUnit.Type),
		// Обратная связь 1:М к Зданиям (Buildings)
		edge.To("buildings", Building.Type),
	}
}
