package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
    "entgo.io/ent/schema/edge"
	"time"
)

// Checklist holds the schema definition for the Checklist entity.
type Checklist struct {
	ent.Schema
}

// Определяем список допустимых значений для типа осмотра.
var InspectionTypes = []string{"spring", "winter", "partial"}

// Fields of the Checklist.
func (Checklist) Fields() []ent.Field {
	return []ent.Field{
		field.String("title"). Unique(),

		field.Enum("inspection_type").
			Values(InspectionTypes...).
			Comment("Тип осмотра: весенний, зимний или частичный.").
			Default("partial"), // Устанавливаем значение по умолчанию

		// НОВОЕ: Описание
		field.Text("description").
			Optional(),
			
		// НОВОЕ: created_at
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
	}
} 

// Edges of the Checklist.
func (Checklist) Edges() []ent.Edge {
	return []ent.Edge{
		// Связь "многие ко многим" с ElementCatalog через связующую таблицу ChecklistElement.
        edge.To("elements", ChecklistElement.Type),
        
        // Обратная связь: один чек-лист может быть использован во многих Заданиях.
        edge.To("tasks", Task.Type),
	}
}
