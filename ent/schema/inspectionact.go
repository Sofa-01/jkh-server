package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
    "entgo.io/ent/schema/edge"
	"time"
)

// InspectionAct holds the schema definition for the InspectionAct entity.
type InspectionAct struct {
	ent.Schema
}

// Fields of the InspectionAct.
func (InspectionAct) Fields() []ent.Field {
	return []ent.Field{
		field.Int("task_id").
			Unique(),

		field.Time("created_at").
			Default(time.Now).
			Immutable(),
			
		field.Time("approved_at").
			Optional(), // Может быть NULL
			
		field.String("status").
			Default("создан"),
			
		field.Text("conclusion").
			Optional(),
			
		field.String("document_path"). // Путь к сгенерированному PDF
			MaxLen(500).
			Optional(),
	}
}

// Edges of the InspectionAct.
func (InspectionAct) Edges() []ent.Edge {
	return []ent.Edge{
		// Связь 1:1 к Заданию (Task). ФК task_id UNIQUE
		edge.From("task", Task.Type).
			Ref("act"). // Обратная ссылка в Task
			Unique().
			Required().
			Field("task_id"), // Явно указываем имя ФК
	}
}
