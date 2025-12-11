// pkg/server/router.go

package server

import (
	"jkh/ent"
	"jkh/pkg/handlers"
	"jkh/pkg/middleware"
	"jkh/pkg/service"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func SetupRouter(client *ent.Client) *gin.Engine {
	//создаёт движок Gin и включает стандартные middleware (логирование и обработку паник)
	r := gin.Default()

	// Swagger UI — документация API доступна по адресу /swagger/index.html
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// --- ИНИЦИАЛИЗАЦИЯ СЕРВИСОВ И ХЕНДЛЕРОВ ---
	authHandler := handlers.NewAuthHandler(client)

	userService := service.NewUserService(client)
	userHandler := handlers.NewUserHandler(userService)

	districtService := service.NewDistrictService(client)
	districtHandler := handlers.NewDistrictHandler(districtService)

	jkhUnitService := service.NewJkhUnitService(client)
	jkhUnitHandler := handlers.NewJkhUnitHandler(jkhUnitService)

	buildingService := service.NewBuildingService(client)
	buildingHandler := handlers.NewBuildingHandler(buildingService)

	elementCatalogService := service.NewElementCatalogService(client)
	elementCatalogHandler := handlers.NewElementCatalogHandler(elementCatalogService)

	checklistService := service.NewChecklistService(client)
	checklistHandler := handlers.NewChecklistHandler(checklistService)

	taskService := service.NewTaskService(client)
	taskHandler := handlers.NewTaskHandler(taskService)

	inspectionResultService := service.NewInspectionResultService(client)
	inspectionResultHandler := handlers.NewInspectionResultHandler(inspectionResultService)

	// InspectionAct (PDF generation)
	inspectionActService := service.NewInspectionActService(client, "storage/acts")
	inspectionActHandler := handlers.NewInspectionActHandler(inspectionActService)

	// InspectorUnit service/handler (assign inspectors to JKH units)
	inspectorUnitService := service.NewInspectorUnitService(client)
	inspectorUnitHandler := handlers.NewInspectorUnitHandler(inspectorUnitService)

	// Аналитика (preview и генерация PDF)
	analyticsService := service.NewAnalyticsService(client)
	analyticsHandler := handlers.NewAnalyticsHandler(analyticsService)

	v1 := r.Group("/api/v1")
	{
		// --- 1. ПУБЛИЧНЫЕ МАРШРУТЫ (БЕЗ ТОКЕНА) ---
		auth := v1.Group("/auth")
		{
			auth.POST("/login", authHandler.Login)
		}

		// --- 2. ЗАЩИЩЁННЫЕ МАРШРУТЫ ---
		protected := v1.Group("/")
		protected.Use(middleware.AuthRequired())

		// --- A. Администратор / Специалист ---
		specialist := protected.Group("/admin")
		specialist.Use(middleware.RBACMiddleware(middleware.RoleSpecialist))
		{
			// пользователи
			specialist.POST("/users", userHandler.CreateUser)
			specialist.GET("/users", userHandler.ListUsers)
			specialist.GET("/users/:id", userHandler.GetUser)
			specialist.PUT("/users/:id", userHandler.UpdateUser)
			specialist.DELETE("/users/:id", userHandler.DeleteUser)

			specialist.POST("/districts", districtHandler.CreateDistrict)
			specialist.GET("/districts", districtHandler.ListDistricts)
			specialist.GET("/districts/:id", districtHandler.GetDistrict)
			specialist.PUT("/districts/:id", districtHandler.UpdateDistrict)
			specialist.DELETE("/districts/:id", districtHandler.DeleteDistrict)

			specialist.POST("/jkhunits", jkhUnitHandler.CreateJkhUnit)
			specialist.GET("/jkhunits", jkhUnitHandler.ListJkhUnits)
			specialist.GET("/jkhunits/:id", jkhUnitHandler.GetJkhUnit)
			specialist.PUT("/jkhunits/:id", jkhUnitHandler.UpdateJkhUnit)
			specialist.DELETE("/jkhunits/:id", jkhUnitHandler.DeleteJkhUnit)

			// Управление назначениями инспекторов на ЖЭУ
			specialist.POST("/jkhunits/:id/inspectors", inspectorUnitHandler.AssignInspector)
			specialist.GET("/jkhunits/:id/inspectors", inspectorUnitHandler.ListInspectorsForUnit)
			specialist.DELETE("/jkhunits/:id/inspectors/:inspector_id", inspectorUnitHandler.UnassignInspector)
			// Список ЖЭУ для инспектора
			specialist.GET("/users/:id/jkhunits", inspectorUnitHandler.ListUnitsForInspector)

			specialist.POST("/buildings", buildingHandler.CreateBuilding)
			specialist.GET("/buildings", buildingHandler.ListBuildings)
			specialist.GET("/buildings/:id", buildingHandler.GetBuilding)
			specialist.PUT("/buildings/:id", buildingHandler.UpdateBuilding)
			specialist.DELETE("/buildings/:id", buildingHandler.DeleteBuilding)

			specialist.POST("/elements", elementCatalogHandler.CreateElement)
			specialist.GET("/elements", elementCatalogHandler.ListElements)
			specialist.GET("/elements/:id", elementCatalogHandler.GetElement)
			specialist.PUT("/elements/:id", elementCatalogHandler.UpdateElement)
			specialist.DELETE("/elements/:id", elementCatalogHandler.DeleteElement)

			// для чек-листов
			specialist.POST("/checklists", checklistHandler.CreateChecklist)
			specialist.GET("/checklists", checklistHandler.ListChecklists)
			specialist.GET("/checklists/:id", checklistHandler.GetChecklist)
			specialist.PUT("/checklists/:id", checklistHandler.UpdateChecklist)
			specialist.DELETE("/checklists/:id", checklistHandler.DeleteChecklist)
			// Управление элементами в чек-листах
			specialist.POST("/checklists/:id/elements", checklistHandler.AddElementToChecklist)
			specialist.DELETE("/checklists/:id/elements/:element_id", checklistHandler.RemoveElementFromChecklist)
			specialist.PUT("/checklists/:id/elements/:element_id", checklistHandler.UpdateElementOrder)

			specialist.DELETE("/tasks/:id", taskHandler.DeleteTask)

		}

		// --- B. Координатор ---
		coordinator := protected.Group("/tasks")
		coordinator.Use(middleware.RBACMiddleware(middleware.RoleCoordinator))
		{
			coordinator.POST("/", taskHandler.CreateTask)                // Создать задание
			coordinator.GET("/", taskHandler.ListAllTasks)               // Список всех заданий
			coordinator.GET("/:id", taskHandler.GetTask)                 // Детали задания
			coordinator.PUT("/:id/status", taskHandler.UpdateTaskStatus) // Изменить статус
			coordinator.PUT("/:id/assign", taskHandler.AssignInspector)  // Переназначить инспектора

			coordinator.GET("/analytics/preview", analyticsHandler.PreviewChart)
			coordinator.POST("/analytics/report", analyticsHandler.GenerateReport)
		}

		// --- C. Инспектор ---
		inspector := protected.Group("/inspector")
		inspector.Use(middleware.RBACMiddleware(middleware.RoleInspector))
		{
			inspector.GET("/tasks", taskHandler.ListMyTasks)            // Мои задания
			inspector.GET("/tasks/:id", taskHandler.GetTask)            // Детали задания
			inspector.POST("/tasks/:id/accept", taskHandler.AcceptTask) // Принять задание
			inspector.POST("/tasks/:id/submit", taskHandler.SubmitTask) // Отправить на проверку

			inspector.POST("/tasks/:id/results", inspectionResultHandler.CreateOrUpdateResult)       //Создать/обновить результат проверки
			inspector.GET("/tasks/:id/results", inspectionResultHandler.GetTaskResults)              //Получить все результаты задания
			inspector.DELETE("/tasks/:id/results/:element_id", inspectionResultHandler.DeleteResult) //Удалить результат

			inspector.GET("/tasks/:id/act", inspectionActHandler.DownloadAct) //Скачивание акта осмотра (PDF)
		}
	}

	return r
}
