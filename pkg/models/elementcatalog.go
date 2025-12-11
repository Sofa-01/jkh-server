package models

// CreateElementCatalogRequest — DTO для входящих запросов на создание/обновление элемента справочника.
// Используется как в POST (создание), так и в PUT (обновление) запросах.
type CreateElementCatalogRequest struct {
    // Название элемента (например, "Фундамент", "Кровля", "Стены").
    // Поле обязательно (binding:"required") — валидация на уровне Gin.
    Name string `json:"name" binding:"required"`
    
    // Категория элемента (опционально, для фильтрации и группировки).
    // Указатель (*string) позволяет отличить "не передано" от "пустая строка".
    // omitempty — если поле nil, оно не включается в JSON-ответ.
    Category *string `json:"category,omitempty"` // nullable
}

// ElementCatalogResponse — DTO для исходящих ответов (возвращаем клиенту).
// Формат данных оптимизирован под потребности фронтенда.
type ElementCatalogResponse struct {
    ID       int    `json:"id"`       // Уникальный идентификатор элемента
    Name     string `json:"name"`     // Название элемента
    Category string `json:"category"` // Категория (всегда строка, даже если пустая)
}
