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
	"strings"
	"testing"

	"context"

	"cloud.google.com/go/storage"
	"github.com/gin-gonic/gin"
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
		fmt.Println(r)
		arrR := strings.Split(r, "-")
		var regex, _ = regexp.Compile(`[\D]`)
		abc := regex.ReplaceAll([]byte(arrR[0]), []byte(""))

		a, _ := os.Open("8Vg1pnY54eZHUkCHfC6EV4WMpoJSJesZ499.mp4")
		tmpR, _ := strconv.ParseInt(string(abc), 10, 64)
		if tmpR == 1 {
			tmpR = 0
		}
		// if tmpR > 100000 && tmpR < 1000000 {
		// 	fmt.Println("CHANGED!")
		// 	a, _ = os.Open("test copy.mp4")
		// }

		stat, _ := a.Stat()
		exp := tmpR % 8
		if exp == 0 {
			exp = 1
		}
		chunk := math.Pow10(int(exp))
		start := tmpR

		end := math.Min(float64(start)+chunk, float64(stat.Size())-1)
		if len(arrR) > 1 && arrR[1] != "" {
			end, _ = strconv.ParseFloat(arrR[1], 64)
			chunk = end - float64(start)
		}

		koko := io.NewSectionReader(a, int64(start), int64(chunk))

		fmt.Println(fmt.Sprintf("bytes %v-%v/%v", start, int(end), stat.Size()))

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
		// rd, _ := client.Bucket("mindtera-video-dev").
		// Object("5BPltZkKok29UFXzW0I1XJDxHfyEJlxO614.mp4").
		NewRangeReader(context.Background(), start, int64(chunk))
		defer rd.Close()

		end := math.Min(float64(start)+chunk, float64(rd.Size())-1)

		fmt.Println(fmt.Sprintf("bytes %v-%v/%v", start, int(end), rd.Size()))

		c.Stream(func(w io.Writer) bool {
			if int64(end) < (rd.Size() - 1) {
				// c.DataFromReader(
				// 	http.StatusOK,
				// 	int64(chunk), "video/mp4",
				// 	rd, map[string]string{
				// 		"Accept-Ranges": "bytes",
				// 		"Content-Range": fmt.Sprintf("bytes %v-%v/%v", start, int(end), rd.Size()),
				// 	})
				c.SSEvent("s", "")
				return true
			}

			return false
		})
	})

	router.Run()
}

func TestSegment(t *testing.T) {
	inputFile := "1_tmp.mp4"
	baseDir := "test"
	indexFileName := "playlist.m3u8"
	baseFileName := "mongotv"
	baseFileExtension := ".ts"
	segmentLength := 5
	maxSegments := 2048
	// init ffmpeg avformat avcodec register
	Init()

	// sement mp4 file to m3u8
	if err := Segment(inputFile, baseDir, indexFileName, baseFileName, baseFileExtension, segmentLength, maxSegments); err != nil {
		t.Errorf("segment file=%s faild.\n", inputFile)
	}
}
