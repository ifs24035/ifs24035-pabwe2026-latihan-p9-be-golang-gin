package server

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm"

	"app/docs"
	"app/internal/config"
	"app/internal/features/admin"
	"app/internal/features/ai"
	"app/internal/features/auth"
	"app/internal/features/todo"
	"app/internal/middleware"
	"app/internal/response"
)

func SetupRouter(cfg *config.Config, db *gorm.DB) *gin.Engine {
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.CORS())
	r.Use(middleware.RateLimiter(1000, time.Second))

	// Repositories & Services
	userRepo := auth.NewUserRepository(db)
	authSvc := auth.NewAuthService(userRepo, cfg.AppKey)
	authHdr := auth.NewAuthHandler(authSvc)

	todoRepo := todo.NewTodoRepository(db)
	todoSvc := todo.NewTodoService(todoRepo)
	todoHdr := todo.NewTodoHandler(todoSvc)

	searchClient := ai.NewSearchEngineClient(cfg.SearchEngineEndpoint)
	aiClient := ai.NewOpenAiClient(cfg.OpenAIEndpoint, cfg.OpenAIKey, cfg.OpenAIModel)
	aiSvc := ai.NewAiService(aiClient, searchClient)
	aiHdr := ai.NewAiHandler(aiSvc)

	adminSvc := admin.NewAdminService(db)
	adminHdr := admin.NewAdminHandler(adminSvc)

	// Helper to get dynamic swagger JSON adjusted to the current request URL
	getSwaggerData := func(c *gin.Context) []byte {
		host := c.GetHeader("X-Forwarded-Host")
		if host == "" {
			host = c.Request.Host
		}

		scheme := "http"
		if c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https" {
			scheme = "https"
		}

		return docs.GetSwaggerJSON(host, scheme)
	}

	// Redirect /swagger to /swagger/index.html
	r.GET("/swagger", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/swagger/index.html")
	})

	// Swagger Documentation Route (points to api.json and adjusts dynamically to current URL)
	swaggerHandler := ginSwagger.WrapHandler(swaggerFiles.Handler, ginSwagger.URL("/swagger/api.json"))
	r.GET("/swagger/*any", func(c *gin.Context) {
		anyParam := c.Param("any")
		if anyParam == "" || anyParam == "/" {
			c.Redirect(http.StatusMovedPermanently, "/swagger/index.html")
			return
		}
		if anyParam == "/api.json" || anyParam == "/doc.json" {
			c.Data(http.StatusOK, "application/json; charset=utf-8", getSwaggerData(c))
			return
		}
		swaggerHandler(c)
	})

	// Direct endpoint for api.json at root level (adjusts dynamically to current URL)
	r.GET("/api.json", func(c *gin.Context) {
		c.Data(http.StatusOK, "application/json; charset=utf-8", getSwaggerData(c))
	})

	api := r.Group("/api")
	{
		// Auth Routes (Public)
		authRoutes := api.Group("/auth")
		{
			authRoutes.POST("/register", authHdr.Register)
			authRoutes.POST("/login", authHdr.Login)
			authRoutes.POST("/logout", authHdr.Logout)

			// Protected Profile Routes
			authProtected := authRoutes.Group("")
			authProtected.Use(middleware.Auth(cfg.AppKey))
			{
				authProtected.GET("/profile", authHdr.GetProfile)
				authProtected.PUT("/profile", authHdr.UpdateProfile)
			}
		}

		// Public Template Routes
		api.GET("/templates", todoHdr.GetPublicTemplates)
		api.GET("/templates/:id", todoHdr.GetPublicTemplateByID)

		// Protected Routes
		protected := api.Group("")
		protected.Use(middleware.Auth(cfg.AppKey))
		{
			// Clone template
			protected.POST("/templates/:id/clone", todoHdr.CloneTemplate)

			// Todos
			protected.GET("/todos", todoHdr.GetTodos)
			protected.POST("/todos", todoHdr.CreateTodo)
			protected.GET("/todos/stats", todoHdr.GetStats)
			protected.GET("/todos/:id", todoHdr.GetTodoByID)
			protected.PUT("/todos/:id", todoHdr.UpdateTodo)
			protected.PATCH("/todos/:id/status", todoHdr.UpdateTodoStatus)
			protected.DELETE("/todos/:id", todoHdr.DeleteTodo)

			// Groups
			protected.GET("/todo-groups", todoHdr.GetGroups)
			protected.POST("/todo-groups", todoHdr.CreateGroup)
			protected.GET("/todo-groups/:id", todoHdr.GetGroupByID)
			protected.PUT("/todo-groups/:id", todoHdr.UpdateGroup)
			protected.DELETE("/todo-groups/:id", todoHdr.DeleteGroup)

			// AI Todo Generator
			protected.POST("/todos/ai-generate", aiHdr.GenerateTodos)

			// Admin Routes
			adm := protected.Group("/admin")
			adm.Use(middleware.RequireRole("admin"))
			{
				adm.GET("/users", adminHdr.GetAllUsers)
				adm.GET("/users/:id/todos", adminHdr.GetUserTodos)
			}
		}
	}

	// Pure REST API 404 handler for unmatched routes
	r.NoRoute(func(c *gin.Context) {
		response.Fail(c, http.StatusNotFound, "Endpoint not found")
	})

	return r
}
