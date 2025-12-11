package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
    "entgo.io/ent/schema/edge"
	"time"
)
// InspectionResult holds the schema definition for the InspectionResult entity.
type InspectionResult struct {
	ent.Schema
}

var ConditionStatuses = []string{"Исправное", "Удовлетворительное", "Неудовлетворительное", "Аварийное"}

// Fields of the InspectionResult.
func (InspectionResult) Fields() []ent.Field {
	return []ent.Field{
		// 1. ЯВНОЕ ПОЛЕ ФК (ссылка на Task)
		field.Int("task_id"),

		// 2. ЯВНОЕ ПОЛЕ ФК (ссылка на ChecklistElement)
		field.Int("checklist_element_id"),

		// Выбор статуса состояния элемента
        field.Enum("condition_status").
            Values(ConditionStatuses...).
            Comment("Статус состояния: Исправное, Удовлетворительное, Неудовлетворительное, Аварийное."),
            
        // Комментарий инспектора к элементу
        field.String("comment").
            Optional(),
            
        field.Time("created_at").
            Default(time.Now).
            Immutable(),
            
        field.Time("updated_at").
            Default(time.Now).
            UpdateDefault(time.Now),
	}
}

// Edges of the InspectionResult.
// Композитный ключ (task_id + checklist_element_id)
func (InspectionResult) Edges() []ent.Edge {
	return []ent.Edge{
		// FK: Связь с Заданием (Task). (Часть композитного ключа)
        edge.From("task", Task.Type).
            Ref("results").
            Unique().
            Required().
            Field("task_id"),

        // FK: Связь с конкретным элементом в рамках чек-листа (ChecklistElement). (Часть композитного ключа)
        edge.From("checklist_element", ChecklistElement.Type).
            Ref("inspection_results").
            Unique().
            Required().
            Field("checklist_element_id"),
	}
}
