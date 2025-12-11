package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
    "entgo.io/ent/schema/edge"
	"time"
)

// Task holds the schema definition for the Task entity.
type Task struct {
	ent.Schema
}

var Statuses = []string{"New", "Pending", "InProgress", "OnReview", "ForRevision", "Approved", "Canceled"}

// Fields of the Task.
func (Task) Fields() []ent.Field {
	return []ent.Field{
        // 1. ЯВНОЕ ПОЛЕ ФК (ссылка на Building)
		field.Int("building_id"), // <-- ДОБАВЛЕНО
		
		// 2. ЯВНОЕ ПОЛЕ ФК (ссылка на Checklist)
		field.Int("checklist_id"), // <-- ДОБАВЛЕНО
		
		// 3. ЯВНОЕ ПОЛЕ ФК (ссылка на Inspector/User)
		field.Int("inspector_id"), // <-- ДОБАВЛЕНО

		// НОВОЕ: Добавлено поле title
		field.String("title"),
			
		// НОВОЕ: Добавлено поле priority
		field.String("priority").
			Default("обычный"), 
			
		field.Enum("status").
			Values(Statuses...).
			Default("New"), // Сохраняем наше FSM-имя 'New'
			
		field.Text("description"). // НОВОЕ
			Optional(),
			
		field.Time("scheduled_date").
			Comment("Планируемая дата и время осмотра."),
			
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
			
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

// Edges of the Task.
func (Task) Edges() []ent.Edge {
	return []ent.Edge{
		// FK: Назначенный инспектор (связь с User)
		edge.From("inspector", User.Type).
			Ref("inspections").
			Unique().
			Required().
			Field("inspector_id"), // <-- ИЗМЕНЕНО с user_inspections
			
		// FK: Обслуживаемое здание (связь с Building)
		edge.From("building", Building.Type).
			Ref("tasks").
			Unique().
			Required().
			Field("building_id"), // <-- ИЗМЕНЕНО с building_tasks
			
		// FK: Используемый чек-лист (связь с Checklist)
		edge.From("checklist", Checklist.Type).
			Ref("tasks").
			Unique().
			Required().
			Field("checklist_id"), // <-- ИЗМЕНЕНО с checklist_tasks
			
		// 1. Обратная связь к результатам осмотра
		edge.To("results", InspectionResult.Type),
		
		// 2. НОВОЕ: Связь 1:1 к Акту (InspectionAct)
		edge.To("act", InspectionAct.Type).
			Unique(),
	}
}
