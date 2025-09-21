package router

import (
	"database/sql"
	"net/http"
	"time"

	"eventos-backend/internal/domain/checkin"
	"eventos-backend/internal/domain/checkout"
	"eventos-backend/internal/domain/employee"
	"eventos-backend/internal/domain/event"
	"eventos-backend/internal/domain/partner"
	"eventos-backend/internal/domain/permission"
	"eventos-backend/internal/domain/role"
	"eventos-backend/internal/domain/tenant"
	"eventos-backend/internal/domain/user"
	jwtService "eventos-backend/internal/infrastructure/auth/jwt"
	"eventos-backend/internal/infrastructure/monitoring"
	"eventos-backend/internal/interfaces/http/handlers"
	"eventos-backend/internal/interfaces/http/middleware"
	"eventos-backend/internal/interfaces/http/responses"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
)

// Router representa o roteador principal da aplicação
type Router struct {
	engine             *gin.Engine
	logger             *zap.Logger
	db                 *sql.DB
	healthCheckHandler *monitoring.HealthCheckHandler
}

// Config contém as configurações do router
type Config struct {
	Logger            *zap.Logger
	DB                *sql.DB
	JWTService        jwtService.Service
	TenantService     tenant.Service
	UserService       user.Service
	EventService      event.Service
	PartnerService    partner.Service
	EmployeeService   employee.Service
	RoleService       role.Service
	PermissionService permission.Service
	CheckinService    checkin.Service
	CheckoutService   checkout.Service
	// RolePermissionService role.RolePermissionService // TODO: Implementar quando Permission Handler estiver pronto
	Debug bool
}

// New cria uma nova instância do router
func New(cfg Config) *Router {
	// Configurar modo do Gin
	if cfg.Debug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// Criar engine do Gin
	engine := gin.New()

	// Criar handler de health check
	healthCheckHandler := monitoring.NewHealthCheckHandler()

	router := &Router{
		engine:             engine,
		logger:             cfg.Logger,
		db:                 cfg.DB,
		healthCheckHandler: healthCheckHandler,
	}

	// Configurar middleware global
	router.setupMiddleware()

	// Configurar rotas
	router.setupRoutes(cfg)

	return router
}

// Engine retorna o engine do Gin
func (r *Router) Engine() *gin.Engine {
	return r.engine
}

// setupMiddleware configura os middleware globais
func (r *Router) setupMiddleware() {
	// Recovery middleware
	r.engine.Use(middleware.ErrorHandlerMiddleware(r.logger))

	// CORS middleware
	r.engine.Use(middleware.CORSMiddleware())

	// Logging middleware
	r.engine.Use(middleware.LoggingMiddleware(r.logger))

	// Rate limiting middleware
	r.engine.Use(middleware.RateLimiterMiddleware())

	// Request timeout middleware
	r.engine.Use(middleware.TimeoutMiddleware(30 * time.Second))

	// Tracing middleware
	r.engine.Use(middleware.TracingMiddleware())
}

// setupRoutes configura todas as rotas da aplicação
func (r *Router) setupRoutes(cfg Config) {
	// Rotas básicas (sem autenticação)
	r.setupBasicRoutes()

	// Grupo de rotas da API v1
	v1 := r.engine.Group("/api/v1")
	{
		// Rotas de autenticação (sem middleware de auth)
		r.setupAuthRoutes(v1, cfg)

		// Rotas protegidas (com middleware de auth)
		authMiddleware := middleware.NewAuthMiddleware(cfg.JWTService, r.logger)
		protected := v1.Group("")
		protected.Use(authMiddleware.RequireAuth())
		{
			r.setupTenantRoutes(protected, cfg)
			r.setupUserRoutes(protected, cfg)
			r.setupEventRoutes(protected, cfg)
			r.setupPartnerRoutes(protected, cfg)
			r.setupEmployeeRoutes(protected, cfg)
			r.setupRoleRoutes(protected, cfg)
			r.setupPermissionRoutes(protected, cfg)
			r.setupCheckinRoutes(protected, cfg)
			r.setupCheckoutRoutes(protected, cfg)
		}
	}
}

// setupBasicRoutes configura rotas básicas (health, info, etc.)
func (r *Router) setupBasicRoutes() {
	// Health checks
	r.engine.GET("/health", r.healthCheckHandler.HealthCheck)
	r.engine.GET("/ready", r.healthCheckHandler.ReadinessCheck)
	r.engine.GET("/live", r.healthCheckHandler.LivenessCheck)
	r.engine.GET("/metrics", r.healthCheckHandler.Metrics)

	// Documentação Swagger
	r.engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Informações da API
	r.engine.GET("/", r.apiInfo)

	// Ping básico
	r.engine.GET("/ping", func(c *gin.Context) {
		responses.Success(c, gin.H{"message": "pong"}, "")
	})
}

// setupAuthRoutes configura rotas de autenticação
func (r *Router) setupAuthRoutes(rg *gin.RouterGroup, cfg Config) {
	authHandler := handlers.NewAuthHandler(cfg.UserService, cfg.JWTService, r.logger)

	auth := rg.Group("/auth")
	{
		auth.POST("/login", authHandler.Login)
		auth.POST("/refresh", authHandler.RefreshToken)

		// Rotas que precisam de autenticação
		authMiddleware := middleware.NewAuthMiddleware(cfg.JWTService, r.logger)
		auth.POST("/logout", authMiddleware.RequireAuth(), authHandler.Logout)
		auth.GET("/me", authMiddleware.RequireAuth(), authHandler.Me)
	}
}

// setupTenantRoutes configura rotas de tenant
func (r *Router) setupTenantRoutes(rg *gin.RouterGroup, cfg Config) {
	tenantHandler := handlers.NewTenantHandler(cfg.TenantService, r.logger)

	tenants := rg.Group("/tenants")
	{
		tenants.POST("", tenantHandler.Create)
		tenants.GET("/:id", tenantHandler.GetByID)
		tenants.PUT("/:id", tenantHandler.Update)
		tenants.DELETE("/:id", tenantHandler.Delete)
		tenants.GET("", tenantHandler.List)
	}
}

// setupUserRoutes configura rotas de usuário
func (r *Router) setupUserRoutes(rg *gin.RouterGroup, cfg Config) {
	userHandler := handlers.NewUserHandler(cfg.UserService, r.logger)

	users := rg.Group("/users")
	{
		users.POST("", userHandler.Create)
		users.GET("/:id", userHandler.GetByID)
		users.PUT("/:id", userHandler.Update)
		users.DELETE("/:id", userHandler.Delete)
		users.GET("", userHandler.List)
		users.PUT("/:id/password", userHandler.ChangePassword)
	}
}

// setupEventRoutes configura rotas de evento
func (r *Router) setupEventRoutes(rg *gin.RouterGroup, cfg Config) {
	eventHandler := handlers.NewEventHandler(cfg.EventService, r.logger)

	events := rg.Group("/events")
	{
		// Operações básicas
		events.POST("", eventHandler.Create)
		events.GET("/:id", eventHandler.GetByID)
		events.PUT("/:id", eventHandler.Update)
		events.DELETE("/:id", eventHandler.Delete)
		events.GET("", eventHandler.List)

		// Operações específicas
		events.GET("/:id/stats", eventHandler.GetStats)
	}
}

// setupPartnerRoutes configura rotas de parceiro
func (r *Router) setupPartnerRoutes(rg *gin.RouterGroup, cfg Config) {
	partnerHandler := handlers.NewPartnerHandler(cfg.PartnerService, r.logger)

	partners := rg.Group("/partners")
	{
		// Operações básicas
		partners.POST("", partnerHandler.Create)
		partners.GET("/:id", partnerHandler.GetByID)
		partners.PUT("/:id", partnerHandler.Update)
		partners.DELETE("/:id", partnerHandler.Delete)
		partners.GET("", partnerHandler.List)

		// Operações específicas
		partners.POST("/:id/login", partnerHandler.Login)
		partners.POST("/:id/password", partnerHandler.ChangePassword)
	}
}

// setupEmployeeRoutes configura rotas de funcionário
func (r *Router) setupEmployeeRoutes(rg *gin.RouterGroup, cfg Config) {
	employeeHandler := handlers.NewEmployeeHandler(cfg.EmployeeService, r.logger)

	employees := rg.Group("/employees")
	{
		// Operações básicas
		employees.POST("", employeeHandler.Create)
		employees.GET("/:id", employeeHandler.GetByID)
		employees.PUT("/:id", employeeHandler.Update)
		employees.DELETE("/:id", employeeHandler.Delete)
		employees.GET("", employeeHandler.List)

		// Operações específicas
		employees.POST("/:id/photo", employeeHandler.UploadPhoto)
		employees.POST("/:id/face", employeeHandler.UpdateFaceEmbedding)
		employees.POST("/:id/recognize", employeeHandler.RecognizeFace)
	}
}

// setupRoleRoutes configura rotas de role
func (r *Router) setupRoleRoutes(rg *gin.RouterGroup, cfg Config) {
	roleHandler := handlers.NewRoleHandler(cfg.RoleService, r.logger)

	roles := rg.Group("/roles")
	{
		// CRUD básico de roles
		roles.POST("", roleHandler.Create)
		roles.GET("/:id", roleHandler.GetByID)
		roles.PUT("/:id", roleHandler.Update)
		roles.DELETE("/:id", roleHandler.Delete)
		roles.GET("", roleHandler.List)
		roles.GET("/system", roleHandler.ListSystem)

		// Ações de ativação/desativação
		roles.POST("/:id/activate", roleHandler.Activate)
		roles.POST("/:id/deactivate", roleHandler.Deactivate)

		// Utilitários
		roles.GET("/available-levels", roleHandler.GetAvailableLevels)
		roles.GET("/suggest-level", roleHandler.SuggestLevel)
	}
}

// setupPermissionRoutes configura rotas de permissão
func (r *Router) setupPermissionRoutes(rg *gin.RouterGroup, cfg Config) {
	permissionHandler := handlers.NewPermissionHandler(cfg.PermissionService, r.logger)

	permissions := rg.Group("/permissions")
	{
		// CRUD básico de permissões
		permissions.POST("", permissionHandler.Create)
		permissions.GET("/:id", permissionHandler.GetByID)
		permissions.PUT("/:id", permissionHandler.Update)
		permissions.DELETE("/:id", permissionHandler.Delete)
		permissions.GET("", permissionHandler.List)

		// Permissões do sistema
		permissions.GET("/system", permissionHandler.ListSystem)
		permissions.POST("/system/init", permissionHandler.InitializeSystemPermissions)

		// Ações de ativação/desativação
		permissions.POST("/:id/activate", permissionHandler.Activate)
		permissions.POST("/:id/deactivate", permissionHandler.Deactivate)
	}
}

// setupCheckinRoutes configura rotas de check-in
func (r *Router) setupCheckinRoutes(rg *gin.RouterGroup, cfg Config) {
	checkinHandler := handlers.NewCheckinHandler(cfg.CheckinService, r.logger)

	checkins := rg.Group("/checkins")
	{
		// Operações básicas
		checkins.POST("", checkinHandler.PerformCheckin)
		checkins.GET("/:id", checkinHandler.GetByID)
		checkins.GET("", checkinHandler.List)

		// Operações específicas
		checkins.POST("/:id/notes", checkinHandler.AddNote)

		// Estatísticas
		checkins.GET("/stats", checkinHandler.GetStats)
		checkins.GET("/recent", checkinHandler.GetRecent)

		// Filtros específicos
		checkins.GET("/employee/:employee_id", checkinHandler.GetByEmployee)
		checkins.GET("/event/:event_id", checkinHandler.GetByEvent)
	}
}

// setupCheckoutRoutes configura rotas de check-out
func (r *Router) setupCheckoutRoutes(rg *gin.RouterGroup, cfg Config) {
	checkoutHandler := handlers.NewCheckoutHandler(cfg.CheckoutService, r.logger)

	checkouts := rg.Group("/checkouts")
	{
		// Operações básicas
		checkouts.POST("", checkoutHandler.PerformCheckout)
		checkouts.GET("/:id", checkoutHandler.GetByID)
		checkouts.GET("", checkoutHandler.List)

		// Operações específicas
		checkouts.POST("/:id/notes", checkoutHandler.AddNote)

		// Estatísticas
		checkouts.GET("/stats", checkoutHandler.GetStats)
		checkouts.GET("/recent", checkoutHandler.GetRecent)

		// Filtros específicos
		checkouts.GET("/employee/:employee_id", checkoutHandler.GetByEmployee)
		checkouts.GET("/event/:event_id", checkoutHandler.GetByEvent)
	}

	// Rotas de sessões de trabalho
	workSessions := rg.Group("/work-sessions")
	{
		workSessions.GET("", checkoutHandler.GetWorkSessions)
		workSessions.GET("/employee/:employee_id", checkoutHandler.GetEmployeeWorkSessions)
		workSessions.GET("/stats", checkoutHandler.GetWorkStats)
	}
}

// healthCheck endpoint de verificação de saúde
func (r *Router) healthCheck(c *gin.Context) {
	// Verificar saúde do banco de dados
	if err := r.db.Ping(); err != nil {
		r.logger.Error("Database health check failed", zap.Error(err))
		responses.Error(c, http.StatusServiceUnavailable, "Database connection failed", "HEALTH_CHECK_FAILED", nil)
		return
	}

	responses.Success(c, gin.H{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
		"version":   "1.0.0",
		"database":  "connected",
	}, "System is healthy")
}

// apiInfo endpoint de informações da API
func (r *Router) apiInfo(c *gin.Context) {
	responses.Success(c, gin.H{
		"name":        "Eventos Backend API",
		"version":     "1.0.0",
		"description": "Sistema de Check-in em Eventos",
		"docs":        "/swagger",
		"endpoints": gin.H{
			"health": "/health",
			"api":    "/api/v1",
		},
	}, "API information")
}

// notImplemented handler temporário para endpoints não implementados
func (r *Router) notImplemented(c *gin.Context) {
	responses.Error(c, http.StatusNotImplemented, "Endpoint not implemented yet", "NOT_IMPLEMENTED", gin.H{
		"method": c.Request.Method,
		"path":   c.Request.URL.Path,
	})
}
