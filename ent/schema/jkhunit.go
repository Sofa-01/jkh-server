package schema

import (
    "entgo.io/ent"
    "entgo.io/ent/schema/field"
    "entgo.io/ent/schema/edge"
)

// JkhUnit holds the schema definition for the JkhUnit entity.
type JkhUnit struct {
	ent.Schema
}

// Fields of the JkhUnit.
func (JkhUnit) Fields() []ent.Field {
	return []ent.Field{
        // 1. ЯВНОЕ ПОЛЕ ФК (ссылка на District)
		field.Int("district_id"), // <-- ДОБАВЛЕНО
		// Название ЖЭУ (например, "ЖЭУ 5" или "Район Северный").
        field.String("name").
            Unique(),
	}
}

// Edges of the JkhUnit.
func (JkhUnit) Edges() []ent.Edge {
	return []ent.Edge{
        // Связь М:1 к Району (District)
		edge.From("district", District.Type).
			Ref("jkh_units").
			Unique().
			Required().
			Field("district_id"),

		// Обратная связь: одно ЖЭУ может иметь много Зданий (Buildings).
        edge.To("buildings", Building.Type),

        // ВОЗВРАЩЕНО: Связь М:М: Список Инспекторов, привязанных к этому ЖЭУ.
        edge.To("assigned_inspectors", InspectorUnit.Type), // <-- ВОЗВРАЩЕНОо
	}
}
