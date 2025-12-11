package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
    "entgo.io/ent/schema/edge"
)

// User holds the schema definition for the User entity.
type User struct {
	ent.Schema
}

// Fields of the User.
func (User) Fields() []ent.Field {
	return []ent.Field{
        field.Int("role_id"),

		// Уникальный email для входа (логин).
        field.String("email").
            Unique(), 
        
        field.String("login").
            Unique(),
        
        // Хешированный пароль. Мы будем хранить Bcrypt хеш, а не сам пароль. [1]
        field.String("password_hash"). 
            Sensitive(), // Помечаем как чувствительное поле

        // Имя и Фамилия пользователя
        field.String("first_name"),
        field.String("last_name"),
	}
}

// Edges of the User.
func (User) Edges() []ent.Edge {
	return []ent.Edge{
		// Связь М:1 к Роли (Роль переименована в role_id для соответствия SQL)
		edge.From("role", Role.Type).
			Ref("users"). 
			Unique().     
			Required().
			Field("role_id"), // <-- ИЗМЕНЕНО с role_users

		// Обратная связь к Заданиям Инспектора
		edge.To("inspections", Task.Type),

		// Обратная связь к зданиям, назначенным Инспектору (inspector_id в buildings)
		edge.To("assigned_buildings", Building.Type), // <-- НОВАЯ ОБРАТНАЯ СВЯЗЬ

        // ВОЗВРАЩЕНО: Связь М:М: Список ЖЭУ, за которые отвечает пользователь.
        edge.To("assigned_units", InspectorUnit.Type), // <-- ВОЗВРАЩЕНО
	}
}
