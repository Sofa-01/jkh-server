package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
    "entgo.io/ent/schema/edge"
)

// ChecklistElement holds the schema definition for the ChecklistElement entity.
type ChecklistElement struct {
	ent.Schema
}

// Fields of the ChecklistElement.
func (ChecklistElement) Fields() []ent.Field {
	return []ent.Field{
		// Порядок, в котором элемент должен отображаться в чек-листе.
        // Зависит от композитного ключа (checklist_id + element_id).
        // 1. ЯВНОЕ ПОЛЕ ФК (ссылка на Checklist)
		field.Int("checklist_id"), 
		
		// 2. ЯВНОЕ ПОЛЕ ФК (ссылка на ElementCatalog)
		field.Int("element_id"),

        field.Int("order_index").
            Optional(),
	}
}

// Edges of the ChecklistElement.
func (ChecklistElement) Edges() []ent.Edge {
	return []ent.Edge{
		// Связь с Чек-листом (FK part 1)
        edge.From("checklist", Checklist.Type).
            Ref("elements").
            Unique().
            Required().
            Field("checklist_id"), // Явно указываем имя FK 
            
        // Связь со Справочником Элементов (FK part 2)
        edge.From("element_catalog", ElementCatalog.Type).
            Ref("checklist_elements"). 
            Unique().
            Required().
            Field("element_id"), 
            
        // Обратная связь к Акту осмотра (InspectionResult)
        edge.To("inspection_results", InspectionResult.Type),
	}
}
