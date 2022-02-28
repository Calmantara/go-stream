package main

import "cloud.google.com/go/storage"

type CloudStorageParam struct {
	serviceAccountPath string
	objectName         string
	bucketName         string
	client             *storage.Client
}
