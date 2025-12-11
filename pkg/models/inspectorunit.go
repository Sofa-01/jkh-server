// pkg/models/inspectorunit.go

package models

// AssignInspectorToUnitRequest — DTO для назначения инспектора на ЖЭУ
type AssignInspectorToUnitRequest struct {
	InspectorID int `json:"inspector_id" binding:"required,min=1"`
}

// InspectorAssignmentResponse — ответ при успешном создании назначения
type InspectorAssignmentResponse struct {
	JkhUnitID   int    `json:"jkh_unit_id"`
	InspectorID int    `json:"inspector_id"`
	Message     string `json:"message,omitempty"`
}
