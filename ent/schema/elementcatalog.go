package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
    "entgo.io/ent/schema/edge"
)

// ElementCatalog holds the schema definition for the ElementCatalog entity.
type ElementCatalog struct {
	ent.Schema
}

// Fields of the ElementCatalog.
func (ElementCatalog) Fields() []ent.Field {
	return []ent.Field{
		// Название элемента (например, "Фундамент", "Кровля").
        field.String("name").
            Unique(),
        
        // Категория элемента (для удобства фильтрации)
        field.String("category").
            Optional(),
	}
}

// Edges of the ElementCatalog.
func (ElementCatalog) Edges() []ent.Edge {
	return []ent.Edge{
		// Здесь будет обратная связь с ChecklistElement
		// Связь с элементами в конкретных чек-листах (M:M)
        edge.To("checklist_elements", ChecklistElement.Type),
	}
}
