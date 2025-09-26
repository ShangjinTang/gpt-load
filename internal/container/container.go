// Package container provides a dependency injection container for the application.
package container

import (
	"gpt-load/internal/app"
	"gpt-load/internal/channel"
	"gpt-load/internal/config"
	"gpt-load/internal/db"
	"gpt-load/internal/encryption"
	"gpt-load/internal/handler"
	"gpt-load/internal/httpclient"
	"gpt-load/internal/interfaces"
	"gpt-load/internal/keypool"
	"gpt-load/internal/policy"
	"gpt-load/internal/proxy"
	"gpt-load/internal/router"
	"gpt-load/internal/services"
	"gpt-load/internal/store"
	"gpt-load/internal/types"
	"gpt-load/internal/validator"

	"go.uber.org/dig"
	"gorm.io/gorm"
)

// BuildContainer creates a new dependency injection container and provides all the application's services.
func BuildContainer() (*dig.Container, error) {
	container := dig.New()

	// Infrastructure Services
	if err := container.Provide(config.NewManager); err != nil {
		return nil, err
	}
	if err := container.Provide(func(configManager types.ConfigManager) (encryption.Service, error) {
		return encryption.NewService(configManager.GetEncryptionKey())
	}); err != nil {
		return nil, err
	}
	if err := container.Provide(db.NewDB); err != nil {
		return nil, err
	}
	if err := container.Provide(config.NewSystemSettingsManager); err != nil {
		return nil, err
	}
	if err := container.Provide(store.NewStore); err != nil {
		return nil, err
	}
	if err := container.Provide(httpclient.NewHTTPClientManager); err != nil {
		return nil, err
	}
	if err := container.Provide(channel.NewFactory); err != nil {
		return nil, err
	}

	// Business Services
	if err := container.Provide(services.NewTaskService); err != nil {
		return nil, err
	}
	if err := container.Provide(services.NewKeyManualValidationService); err != nil {
		return nil, err
	}
	if err := container.Provide(services.NewKeyService); err != nil {
		return nil, err
	}
	if err := container.Provide(services.NewKeyImportService); err != nil {
		return nil, err
	}
	if err := container.Provide(services.NewKeyDeleteService); err != nil {
		return nil, err
	}
	if err := container.Provide(services.NewLogService); err != nil {
		return nil, err
	}
	if err := container.Provide(services.NewLogCleanupService); err != nil {
		return nil, err
	}
	if err := container.Provide(services.NewRequestLogService); err != nil {
		return nil, err
	}
	if err := container.Provide(services.NewGroupManager); err != nil {
		return nil, err
	}
	if err := container.Provide(services.NewKeyStateService); err != nil {
		return nil, err
	}
	if err := container.Provide(func(db *gorm.DB) *policy.PolicyEngine {
		return policy.NewPolicyEngine(db)
	}); err != nil {
		return nil, err
	}
	if err := container.Provide(func(db *gorm.DB) *policy.PolicyService {
		return policy.NewPolicyService(db)
	}); err != nil {
		return nil, err
	}
	if err := container.Provide(services.NewKeyPolicyHandler); err != nil {
		return nil, err
	}
	if err := container.Provide(services.NewIncrementalValidationService); err != nil {
		return nil, err
	}
	// Provide interface implementations
	if err := container.Provide(func(engine *policy.PolicyEngine) interfaces.PolicyEngineInterface {
		return engine
	}); err != nil {
		return nil, err
	}
	if err := container.Provide(func(service *services.KeyStateService) interfaces.KeyStateServiceInterface {
		return service
	}); err != nil {
		return nil, err
	}
	if err := container.Provide(func(handler *services.KeyPolicyHandler) interfaces.KeyPolicyHandlerInterface {
		return handler
	}); err != nil {
		return nil, err
	}
	if err := container.Provide(func(service *services.IncrementalValidationService) interfaces.IncrementalValidationServiceInterface {
		return service
	}); err != nil {
		return nil, err
	}
	if err := container.Provide(keypool.NewProvider); err != nil {
		return nil, err
	}
	if err := container.Provide(validator.NewKeyValidator); err != nil {
		return nil, err
	}
	if err := container.Provide(keypool.NewCronChecker); err != nil {
		return nil, err
	}

	// Handlers
	if err := container.Provide(handler.NewServer); err != nil {
		return nil, err
	}
	if err := container.Provide(handler.NewCommonHandler); err != nil {
		return nil, err
	}
	if err := container.Provide(handler.NewIncrementalValidationHandler); err != nil {
		return nil, err
	}

	// Proxy & Router
	if err := container.Provide(proxy.NewProxyServer); err != nil {
		return nil, err
	}
	if err := container.Provide(router.NewRouter); err != nil {
		return nil, err
	}

	// Application Layer
	if err := container.Provide(app.NewApp); err != nil {
		return nil, err
	}

	return container, nil
}
