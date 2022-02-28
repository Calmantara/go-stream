package main

import "cloud.google.com/go/storage"

type CloudStorageConfiguration interface {
	GetClient() *storage.Client
	GetParam() CloudStorageParam
}

type CloudStorageConfigurationImpl struct {
	cloudStorageParam CloudStorageParam
}
