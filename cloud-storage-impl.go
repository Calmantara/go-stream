package main

import (
	"log"
	"os"

	"context"

	"cloud.google.com/go/storage"
)

func NewCloudStorageConfiguration() CloudStorageConfiguration {
	ServiceAccountPath := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")

	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		log.Fatal(err.Error())
	}

	cloudStorageParam := CloudStorageParam{
		serviceAccountPath: ServiceAccountPath,
		client:             client,
	}

	return &CloudStorageConfigurationImpl{
		cloudStorageParam: cloudStorageParam,
	}
}

func (c *CloudStorageConfigurationImpl) GetClient() *storage.Client {
	return c.cloudStorageParam.client
}
func (c *CloudStorageConfigurationImpl) GetParam() CloudStorageParam {
	return c.cloudStorageParam
}
