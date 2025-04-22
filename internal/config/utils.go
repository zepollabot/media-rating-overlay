package config

import (
	"fmt"
	"runtime"

	"go.uber.org/zap"

	config "github.com/zepollabot/media-rating-overlay/internal/config/model"
	service "github.com/zepollabot/media-rating-overlay/internal/config/service"
)

// LoadConfig loads the application configuration
func LoadConfig(env string) (*config.Config, error) {
	config := service.NewConfigService(service.DefaultEnvConfigDir, service.DefaultConfigFileName, service.DefaultEnvConfigFilePattern, env)
	appConfig, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("error loading configuration: %w", err)
	}
	return appConfig, nil
}

// ConfigureSystemResources configures system resources based on the configuration
func ConfigureSystemResources(perfConfig config.Performance, logger *zap.Logger) int {
	numThreads := runtime.NumCPU()

	// Determine the maximum number of threads to use
	var maxThreads int
	if perfConfig.MaxThreads != 0 {
		if perfConfig.MaxThreads >= numThreads {
			maxThreads = numThreads - 1
		} else {
			maxThreads = perfConfig.MaxThreads
		}
	} else {
		maxThreads = numThreads - 1
	}

	logger.Info("System information", zap.Int("CPU", numThreads))
	logger.Info("Configured maximum number of CPUs", zap.Int("maxThreads", maxThreads))

	runtime.GOMAXPROCS(maxThreads)
	return maxThreads
}
