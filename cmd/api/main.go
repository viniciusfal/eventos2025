// @title Eventos API
// @version 1.0
// @description Sistema de Check-in em Eventos - API REST para gestão de eventos, check-ins e autenticação
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
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
	"eventos-backend/internal/infrastructure/auth/jwt"
	"eventos-backend/internal/infrastructure/cache"
	redisCache "eventos-backend/internal/infrastructure/cache/redis"
	"eventos-backend/internal/infrastructure/config"
	"eventos-backend/internal/infrastructure/messaging/handlers"
	"eventos-backend/internal/infrastructure/messaging/rabbitmq"
	"eventos-backend/internal/infrastructure/persistence/postgres"
	"eventos-backend/internal/infrastructure/persistence/postgres/repositories"
	"eventos-backend/internal/interfaces/http/router"

	"go.uber.org/zap"
)

func main() {
	// Carregar configuração
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Configurar logger
	logger, err := setupLogger(cfg.Logging)
	if err != nil {
		log.Fatalf("Failed to setup logger: %v", err)
	}
	defer logger.Sync()

	logger.Info("Starting eventos-backend application")

	// Configurar banco de dados
	dbConfig := postgres.Config{
		Host:            cfg.Database.Host,
		Port:            cfg.Database.Port,
		Database:        cfg.Database.Name,
		Username:        cfg.Database.User,
		Password:        cfg.Database.Password,
		SSLMode:         cfg.Database.SSLMode,
		MaxOpenConns:    cfg.Database.MaxOpenConns,
		MaxIdleConns:    cfg.Database.MaxIdleConns,
		ConnMaxLifetime: cfg.Database.ConnMaxLifetime,
	}

	db, err := postgres.NewConnection(dbConfig, logger)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	// Configurar Redis Cache
	redisConfig := redisCache.Config{
		Host:            cfg.Redis.Host,
		Port:            cfg.Redis.Port,
		Password:        cfg.Redis.Password,
		DB:              cfg.Redis.DB,
		MaxRetries:      3,
		PoolSize:        cfg.Redis.PoolSize,
		MinIdleConns:    cfg.Redis.MinIdleConns,
		DialTimeout:     5 * time.Second,
		ReadTimeout:     3 * time.Second,
		WriteTimeout:    3 * time.Second,
		IdleTimeout:     300 * time.Second,
		ConnMaxLifetime: time.Hour,
	}

	redisClient, err := redisCache.NewClient(redisConfig, logger)
	if err != nil {
		logger.Warn("Failed to connect to Redis, continuing without cache", zap.Error(err))
		redisClient = nil
	}

	// Configurar Cache Manager e Service
	var cacheService *cache.CacheService
	if redisClient != nil {
		cacheManager := cache.NewDefaultCacheManager(logger)
		cacheManager.SetCache("default", redisClient)

		keyBuilder := cache.NewDefaultKeyBuilder("eventos", 15*time.Minute)
		cacheService = cache.NewCacheService(cacheManager, keyBuilder, logger, 15*time.Minute)

		defer func() {
			if err := cacheService.Close(); err != nil {
				logger.Error("Failed to close cache service", zap.Error(err))
			}
		}()
	}

	// Configurar RabbitMQ
	rabbitConfig := rabbitmq.Config{
		Host:              cfg.RabbitMQ.Host,
		Port:              cfg.RabbitMQ.Port,
		Username:          cfg.RabbitMQ.User,
		Password:          cfg.RabbitMQ.Password,
		VHost:             cfg.RabbitMQ.VHost,
		ConnectionName:    "eventos-backend",
		Heartbeat:         60 * time.Second,
		ConnectionTimeout: 30 * time.Second,
		MaxRetries:        3,
		RetryDelay:        5 * time.Second,
	}

	rabbitClient, err := rabbitmq.NewClient(rabbitConfig, logger)
	if err != nil {
		logger.Warn("Failed to connect to RabbitMQ, continuing without messaging", zap.Error(err))
		rabbitClient = nil
	}

	// Configurar Publisher e Consumer
	var consumer *rabbitmq.Consumer
	var keyBuilder cache.KeyBuilder
	if rabbitClient != nil {
		// Declarar exchanges e filas básicas
		if err := setupRabbitMQTopology(rabbitClient, logger); err != nil {
			logger.Error("Failed to setup RabbitMQ topology", zap.Error(err))
		}

		// Configurar key builder para o consumer
		if cacheService != nil {
			keyBuilder = cache.NewDefaultKeyBuilder("eventos", 15*time.Minute)
		}

		// Configurar Consumer
		consumerConfig := rabbitmq.ConsumerConfig{
			QueueName:           "eventos.checkin.events",
			ConsumerTag:         "eventos-backend-checkin",
			AutoAck:             false,
			PrefetchCount:       10,
			MaxRetries:          3,
			RetryDelay:          5 * time.Second,
			ProcessingTimeout:   30 * time.Second,
			ConcurrentConsumers: 2,
		}
		consumer = rabbitmq.NewConsumer(rabbitClient, consumerConfig, logger)

		// Registrar handlers de mensagem
		if cacheService != nil {
			checkinHandler := handlers.NewCheckinEventHandler(logger, cacheService, keyBuilder)
			consumer.RegisterHandler(rabbitmq.MessageTypeCheckinPerformed, checkinHandler)
			consumer.RegisterHandler(rabbitmq.MessageTypeCheckinValidated, checkinHandler)
			consumer.RegisterHandler(rabbitmq.MessageTypeCheckinInvalid, checkinHandler)
		}

		defer func() {
			if consumer != nil && consumer.IsRunning() {
				if err := consumer.Stop(); err != nil {
					logger.Error("Failed to stop consumer", zap.Error(err))
				}
			}
			if err := rabbitClient.Close(); err != nil {
				logger.Error("Failed to close RabbitMQ client", zap.Error(err))
			}
		}()
	}

	// Configurar JWT Service
	jwtConfig := jwt.Config{
		SecretKey:         cfg.JWT.Secret,
		Expiration:        cfg.JWT.Expiration,
		RefreshExpiration: cfg.JWT.RefreshExpiration,
		Issuer:            "eventos-backend",
	}
	jwtService := jwt.NewJWTService(jwtConfig)

	// Configurar repositórios
	tenantRepo := repositories.NewTenantRepository(db.DB, logger)
	userRepo := repositories.NewUserRepository(db.DB, logger)
	eventRepo := repositories.NewEventRepository(db.DB, logger)
	partnerRepo := repositories.NewPartnerRepository(db.DB, logger)
	employeeRepo := repositories.NewEmployeeRepository(db.DB, logger)
	roleRepo := repositories.NewRoleRepository(db.DB, logger)
	permissionRepo := repositories.NewPermissionRepository(db.DB, logger)
	checkinRepo := repositories.NewCheckinRepository(db.DB, logger)
	checkoutRepo := repositories.NewCheckoutRepository(db.DB, logger)

	// Configurar serviços de domínio
	tenantService := tenant.NewDomainService(tenantRepo, logger)
	userService := user.NewDomainService(userRepo, logger)
	eventService := event.NewDomainService(eventRepo, logger)
	partnerService := partner.NewDomainService(partnerRepo, logger)
	employeeService := employee.NewDomainService(employeeRepo, logger)
	roleService := role.NewService(roleRepo)
	permissionService := permission.NewService(permissionRepo)

	// Configurar serviços de check-in/check-out
	// Nota: Os serviços precisam de StatsRepository, mas por enquanto usaremos nil
	checkinService := checkin.NewService(checkinRepo, nil)    // TODO: Implementar CheckinStatsRepository
	checkoutService := checkout.NewService(checkoutRepo, nil) // TODO: Implementar CheckoutStatsRepository

	// Configurar router
	routerConfig := router.Config{
		Logger:            logger,
		DB:                db.DB.DB, // Acessar o *sql.DB através do sqlx.DB embutido
		JWTService:        jwtService,
		TenantService:     tenantService,
		UserService:       userService,
		EventService:      eventService,
		PartnerService:    partnerService,
		EmployeeService:   employeeService,
		RoleService:       roleService,
		PermissionService: permissionService,
		CheckinService:    checkinService,
		CheckoutService:   checkoutService,
		Debug:             cfg.Logging.Level == "debug",
	}

	appRouter := router.New(routerConfig)

	// Iniciar consumer de mensagens
	if consumer != nil {
		go func() {
			ctx := context.Background()
			if err := consumer.Start(ctx); err != nil {
				logger.Error("Failed to start message consumer", zap.Error(err))
			}
		}()
	}

	// Configurar servidor HTTP
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:      appRouter.Engine(),
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Iniciar servidor em goroutine
	go func() {
		logger.Info("Starting HTTP server",
			zap.String("address", server.Addr),
		)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Aguardar sinal de interrupção
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server exited")
}

// setupRabbitMQTopology configura exchanges e filas do RabbitMQ
func setupRabbitMQTopology(client *rabbitmq.Client, logger *zap.Logger) error {
	// Declarar exchange principal
	if err := client.DeclareExchange("eventos.events", "topic", true, false, nil); err != nil {
		return fmt.Errorf("failed to declare main exchange: %w", err)
	}

	// Declarar filas para diferentes tipos de eventos
	queues := []struct {
		name       string
		routingKey string
	}{
		{"eventos.checkin.events", "checkin.events"},
		{"eventos.checkout.events", "checkout.events"},
		{"eventos.user.events", "user.events"},
		{"eventos.system.events", "system.events"},
		{"eventos.notification.events", "notification.events"},
	}

	for _, q := range queues {
		// Declarar fila
		queue, err := client.DeclareQueue(q.name, true, false, false, nil)
		if err != nil {
			logger.Error("Failed to declare queue", zap.String("queue", q.name), zap.Error(err))
			continue
		}

		// Bind fila ao exchange
		if err := client.BindQueue(queue.Name, q.routingKey, "eventos.events", nil); err != nil {
			logger.Error("Failed to bind queue", zap.String("queue", q.name), zap.Error(err))
			continue
		}

		logger.Info("Queue configured successfully", zap.String("queue", q.name))
	}

	logger.Info("RabbitMQ topology setup completed")
	return nil
}

func setupLogger(cfg config.LoggingConfig) (*zap.Logger, error) {
	var zapConfig zap.Config

	if cfg.Level == "debug" {
		zapConfig = zap.NewDevelopmentConfig()
	} else {
		zapConfig = zap.NewProductionConfig()
	}

	// Configurar nível de log
	switch cfg.Level {
	case "debug":
		zapConfig.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	case "info":
		zapConfig.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	case "warn":
		zapConfig.Level = zap.NewAtomicLevelAt(zap.WarnLevel)
	case "error":
		zapConfig.Level = zap.NewAtomicLevelAt(zap.ErrorLevel)
	default:
		zapConfig.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	}

	return zapConfig.Build()
}
