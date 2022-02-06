package main

import (
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"regexp"
	"strconv"

	"context"

	"github.com/gin-gonic/gin"

	"cloud.google.com/go/storage"
)

type CloudStorageParam struct {
	serviceAccountPath string
	objectName         string
	bucketName         string
	client             *storage.Client
}

type CloudStorageConfiguration interface {
	GetClient() *storage.Client
	GetParam() CloudStorageParam
}

type CloudStorageConfigurationImpl struct {
	cloudStorageParam CloudStorageParam
}

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

func main() {
	gs := NewCloudStorageConfiguration()
	router := gin.Default()

	router.StaticFS("/test", http.Dir("./"))

	router.GET("/video", func(c *gin.Context) {
		// check range
		r := c.GetHeader("range")
		if r == "" {
			c.AbortWithStatusJSON(
				http.StatusBadRequest,
				"Range is required",
			)
			return
		}
		var regex, _ = regexp.Compile(`[\D]`)
		abc := regex.ReplaceAll([]byte(r), []byte(""))

		a, _ := os.Open("test.mp4")
		stat, _ := a.Stat()

		tmpR, _ := strconv.ParseInt(string(abc), 10, 64)
		chunk := math.Pow10(6)
		start := tmpR
		end := math.Min(float64(start)+chunk, float64(stat.Size())-1)
		koko := io.NewSectionReader(a, int64(start), int64(chunk))

		fmt.Println(fmt.Sprintf("bytes %v-%v/%v", start, int(end), stat.Size()))

		io.Pipe()

		c.DataFromReader(
			http.StatusPartialContent,
			int64(chunk), "video/mp4",
			koko, map[string]string{
				"Accept-Ranges": "bytes",
				"Content-Range": fmt.Sprintf("bytes %v-%v/%v", start, int(end), stat.Size()),
			})
	})

	router.GET("/videoplayback", func(c *gin.Context) {

		// check range
		r := c.GetHeader("range")
		if r == "" {
			c.AbortWithStatusJSON(
				http.StatusBadRequest,
				"Range is required",
			)
			return
		}

		var regex, _ = regexp.Compile(`[\D]`)
		abc := regex.ReplaceAll([]byte(r), []byte(""))

		tmpR, _ := strconv.ParseInt(string(abc), 10, 64)
		chunk := 5 * math.Pow10(6)
		start := tmpR

		client := gs.GetClient()
		rd, _ := client.Bucket("bucket_name").
			Object("object_name").
			NewRangeReader(context.Background(), start, int64(chunk))
		defer rd.Close()
		end := math.Min(float64(start)+chunk, float64(rd.Size())-1)

		fmt.Println(fmt.Sprintf("bytes %v-%v/%v", start, int(end), rd.Size()))

		io.Pipe()

		c.DataFromReader(
			http.StatusPartialContent,
			int64(chunk), "video/mp4",
			rd, map[string]string{
				"Accept-Ranges": "bytes",
				"Content-Range": fmt.Sprintf("bytes %v-%v/%v", start, int(end), rd.Size()),
			})
	})

	router.Run()
}
