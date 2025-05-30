package utils

import (
	"carsawa/services/storage"
	"fmt"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/spf13/viper"
)

// LoadCloudinaryConfig loads the Cloudinary configuration from the YAML file.
func LoadCloudinaryConfig() error {
	viper.SetConfigFile("utils/cloudinary.yaml")
	viper.SetConfigType("yaml")

	//fallback opt
	viper.SetDefault("cloudinary.cloudName", "default_cloud_name")
	viper.SetDefault("cloudinary.apiKey", "default_api_key")
	viper.SetDefault("cloudinary.apiSecret", "default_api_secret")
	viper.SetDefault("cloudinary.adminKey", "default_admin_key")

	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("error reading cloudinary config file: %w", err)
	}
	return nil
}

// Cloudinary initializes and returns a Cloudinary-based StorageService using Viper.
func Cloudinary() (storage.StorageService, error) {
	// Ensure the config is loaded.
	if err := LoadCloudinaryConfig(); err != nil {
		return nil, err
	}

	cloudName := viper.GetString("cloudinary.cloudName")
	apiKey := viper.GetString("cloudinary.apiKey")
	apiSecret := viper.GetString("cloudinary.apiSecret")

	if cloudName == "" || apiKey == "" || apiSecret == "" {
		return nil, fmt.Errorf("cloudinary credentials not set in configuration")
	}

	cld, err := cloudinary.NewFromParams(cloudName, apiKey, apiSecret)
	if err != nil {
		return nil, fmt.Errorf("utils.Cloudinary: failed to initialize Cloudinary: %w", err)
	}

	// Create the storage service using our Cloudinary client, cloud name, and apiSecret.
	storageSvc := storage.NewStorageService(cld, cloudName, apiSecret)
	return storageSvc, nil
}
