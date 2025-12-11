package handlers

import (
	"errors"
	"net/http"

	"jkh/ent/task"
	"jkh/pkg/models"
	"jkh/pkg/service"

	"github.com/gin-gonic/gin"
)

// ============================================================================
// ХЕНДЛЕР
// ============================================================================

type TaskHandler struct {
	Service *service.TaskService
}

func NewTaskHandler(s *service.TaskService) *TaskHandler {
	return &TaskHandler{Service: s}
}

// ============================================================================
// CRUD ДЛЯ COORDINATOR/SPECIALIST
// ============================================================================

// CreateTask godoc
// @Summary      Создать задание
// @Description  Создание нового задания на осмотр здания для инспектора
// @Tags         Задания
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body models.CreateTaskRequest true "Данные задания"
// @Success      201 {object} models.TaskDetailResponse "Задание успешно создано"
// @Failure      400 {object} map[string]string "Неверный запрос или FK не найден"
// @Failure      401 {object} map[string]string "Не авторизован"
// @Failure      500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router       /tasks/ [post]
func (h *TaskHandler) CreateTask(c *gin.Context) {
	var req models.CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request or validation failed"})
		return
	}

	resp, err := h.Service.CreateTask(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidForeignKey) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid building, checklist, or inspector ID"})
			return
		}
		if errors.Is(err, service.ErrInspectorNotAssigned) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Inspector is not assigned to this JKH unit"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create task"})
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// ListAllTasks godoc
// @Summary      Получить список всех заданий
// @Description  Возвращает список всех заданий с возможностью фильтрации по статусу
// @Tags         Задания
// @Produce      json
// @Security     BearerAuth
// @Param        status query string false "Фильтр по статусу (New, Pending, InProgress, OnReview, ForRevision, Approved, Canceled)"
// @Success      200 {array} models.TaskResponse "Список заданий"
// @Failure      401 {object} map[string]string "Не авторизован"
// @Failure      500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router       /tasks/ [get]
func (h *TaskHandler) ListAllTasks(c *gin.Context) {
	// Опциональный фильтр по статусу
	var statusFilter *string
	if status := c.Query("status"); status != "" {
		statusFilter = &status
	}

	resp, err := h.Service.ListTasks(c.Request.Context(), nil, statusFilter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve task list"})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// GetTask godoc
// @Summary      Получить задание по ID
// @Description  Возвращает детальную информацию о задании
// @Tags         Задания
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "ID задания"
// @Success      200 {object} models.TaskDetailResponse "Данные задания"
// @Failure      400 {object} map[string]string "Неверный ID"
// @Failure      401 {object} map[string]string "Не авторизован"
// @Failure      404 {object} map[string]string "Задание не найдено"
// @Failure      500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router       /tasks/{id} [get]
func (h *TaskHandler) GetTask(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	resp, err := h.Service.RetrieveTask(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrTaskNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve task"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// UpdateTaskStatus godoc
// @Summary      Изменить статус задания
// @Description  Изменение статуса задания (согласно FSM: New→Pending→InProgress→OnReview→Approved/ForRevision)
// @Tags         Задания
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "ID задания"
// @Param        request body models.UpdateTaskStatusRequest true "Новый статус"
// @Success      200 {object} map[string]string "Статус успешно изменен"
// @Failure      400 {object} map[string]string "Неверный запрос или недопустимый переход статуса"
// @Failure      401 {object} map[string]string "Не авторизован"
// @Failure      404 {object} map[string]string "Задание не найдено"
// @Failure      500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router       /tasks/{id}/status [put]
func (h *TaskHandler) UpdateTaskStatus(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	var req models.UpdateTaskStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request or validation failed"})
		return
	}

	err = h.Service.UpdateTaskStatus(c.Request.Context(), id, task.Status(req.Status))
	if err != nil {
		if errors.Is(err, service.ErrTaskNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
			return
		}
		if errors.Is(err, service.ErrInvalidStatusTransition) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid status transition"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update task status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Task status updated successfully"})
}

// AssignInspector godoc
// @Summary      Назначить инспектора
// @Description  Переназначение инспектора на задание
// @Tags         Задания
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "ID задания"
// @Param        request body models.AssignInspectorRequest true "ID нового инспектора"
// @Success      200 {object} map[string]string "Инспектор успешно назначен"
// @Failure      400 {object} map[string]string "Неверный запрос или инспектор не найден"
// @Failure      401 {object} map[string]string "Не авторизован"
// @Failure      404 {object} map[string]string "Задание не найдено"
// @Failure      500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router       /tasks/{id}/assign [put]
func (h *TaskHandler) AssignInspector(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	var req models.AssignInspectorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request or validation failed"})
		return
	}

	err = h.Service.AssignInspector(c.Request.Context(), id, req.InspectorID)
	if err != nil {
		if errors.Is(err, service.ErrTaskNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
			return
		}
		if errors.Is(err, service.ErrInvalidForeignKey) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid inspector ID"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to assign inspector"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Inspector assigned successfully"})
}

// DeleteTask godoc
// @Summary      Удалить задание
// @Description  Удаление задания из системы
// @Tags         Задания (Администрирование)
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "ID задания"
// @Success      204 "Задание успешно удалено"
// @Failure      400 {object} map[string]string "Неверный ID"
// @Failure      401 {object} map[string]string "Не авторизован"
// @Failure      404 {object} map[string]string "Задание не найдено"
// @Failure      500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router       /admin/tasks/{id} [delete]
func (h *TaskHandler) DeleteTask(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	err = h.Service.DeleteTask(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrTaskNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete task"})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// ============================================================================
// ЭНДПОИНТЫ ДЛЯ INSPECTOR
// ============================================================================

// ListMyTasks godoc
// @Summary      Получить мои задания
// @Description  Возвращает список заданий, назначенных текущему инспектору
// @Tags         Инспектор
// @Produce      json
// @Security     BearerAuth
// @Param        status query string false "Фильтр по статусу"
// @Success      200 {array} models.TaskResponse "Список заданий инспектора"
// @Failure      401 {object} map[string]string "Не авторизован"
// @Failure      500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router       /inspector/tasks [get]
func (h *TaskHandler) ListMyTasks(c *gin.Context) {
	// Извлекаем userID из JWT-токена (установлен middleware AuthRequired)
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	inspectorID := userID.(int)

	// Опциональный фильтр по статусу
	var statusFilter *string
	if status := c.Query("status"); status != "" {
		statusFilter = &status
	}

	resp, err := h.Service.ListTasks(c.Request.Context(), &inspectorID, statusFilter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve task list"})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// AcceptTask godoc
// @Summary      Принять задание
// @Description  Принятие задания инспектором (переход Pending → InProgress)
// @Tags         Инспектор
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "ID задания"
// @Success      200 {object} map[string]string "Задание успешно принято"
// @Failure      400 {object} map[string]string "Неверный ID или недопустимый переход статуса"
// @Failure      401 {object} map[string]string "Не авторизован"
// @Failure      404 {object} map[string]string "Задание не найдено"
// @Failure      500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router       /inspector/tasks/{id}/accept [post]
func (h *TaskHandler) AcceptTask(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	// Переход в статус InProgress
	err = h.Service.UpdateTaskStatus(c.Request.Context(), id, task.StatusInProgress)
	if err != nil {
		if errors.Is(err, service.ErrTaskNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
			return
		}
		if errors.Is(err, service.ErrInvalidStatusTransition) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Task cannot be accepted (invalid status)"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to accept task"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Task accepted successfully"})
}

// SubmitTask godoc
// @Summary      Отправить задание на проверку
// @Description  Отправка выполненного задания на проверку координатору (переход InProgress → OnReview)
// @Tags         Инспектор
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "ID задания"
// @Success      200 {object} map[string]string "Задание отправлено на проверку"
// @Failure      400 {object} map[string]string "Неверный ID или недопустимый переход статуса"
// @Failure      401 {object} map[string]string "Не авторизован"
// @Failure      404 {object} map[string]string "Задание не найдено"
// @Failure      500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router       /inspector/tasks/{id}/submit [post]
func (h *TaskHandler) SubmitTask(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	// TODO: Проверить, что все элементы чек-листа заполнены (InspectionResult)

	// Переход в статус OnReview
	err = h.Service.UpdateTaskStatus(c.Request.Context(), id, task.StatusOnReview)
	if err != nil {
		if errors.Is(err, service.ErrTaskNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
			return
		}
		if errors.Is(err, service.ErrInvalidStatusTransition) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Task cannot be submitted (invalid status)"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to submit task"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Task submitted for review"})
}
