package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
    "entgo.io/ent/schema/edge"
)

// Building holds the schema definition for the Building entity.
type Building struct {
	ent.Schema
}

// Fields of the Building.
func (Building) Fields() []ent.Field {
	return []ent.Field{
        // 1. ЯВНОЕ ПОЛЕ ФК: district_id
		field.Int("district_id"), // <-- ДОБАВЛЕНО
		
		// 2. ЯВНОЕ ПОЛЕ ФК: jkh_unit_id
		field.Int("jkh_unit_id"), // <-- Это уже было добавлено в Edges
		
		// 3. ЯВНОЕ ПОЛЕ ФК: inspector_id (NULLable)
		field.Int("inspector_id").
			Optional(),

		// Уникальный адрес здания
        field.String("address").
            Unique(),

        // Год постройки
        field.Int("construction_year").
            Optional(),

        // НОВОЕ: Описание здания
		field.Text("description").
			Optional(),
			
		// НОВОЕ: Путь к фотографии
		field.String("photo").
			MaxLen(500). // VARCHAR(500)
			Optional(),
	}
}

// Edges of the Building.
func (Building) Edges() []ent.Edge {
	return []ent.Edge{
		// 1. Связь М:1 к ЖЭУ (jkh_unit_id)
		edge.From("jkh_unit", JkhUnit.Type).
			Ref("buildings").
			Unique().
			Required().
			Field("jkh_unit_id"), // <-- ИЗМЕНЕНО с jkh_unit_buildings
			
		// 2. НОВОЕ: Связь М:1 к Району (district_id)
		edge.From("district", District.Type).
			Ref("buildings").
			Unique().
			Required().
			Field("district_id"),
			
		// 3. НОВОЕ: Связь М:1 к Назначенному Инспектору (inspector_id). NULLABLE.
		edge.From("inspector", User.Type).
			Ref("assigned_buildings").
			Unique().
			Field("inspector_id"),

		// Обратная связь к Заданиям.
		edge.To("tasks", Task.Type),
	}
}
