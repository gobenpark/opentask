package task

import (
	"fmt"
	"opentask/pkg/config"
	"opentask/pkg/platforms"
)

func createPlatformClient(platformName string, platform config.Platform) (platforms.PlatformClient, error) {
	// Prepare configuration for platform factory
	clientConfig := make(map[string]any)
	
	// Copy credentials
	for key, value := range platform.Credentials {
		clientConfig[key] = value
	}
	
	// Copy settings
	for key, value := range platform.Settings {
		clientConfig[key] = value
	}
	
	// Create client using registry
	client, err := platforms.DefaultRegistry.Create(platform.Type, clientConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create %s client: %w", platformName, err)
	}
	
	return client, nil
}