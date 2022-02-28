package main

import (
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"

	"context"

	"github.com/gin-gonic/gin"
)

func main() {
	gs := NewCloudStorageConfiguration()
	router := gin.Default()

	router.StaticFS("/page", http.Dir("./"))

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

		a, _ := os.Open("file-name.mp4")
		tmpR, _ := strconv.ParseInt(string(abc), 10, 64)
		if tmpR == 1 {
			tmpR = 0
		}

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
		rd, _ := client.Bucket("bucket-name").
			Object("object-name.mp4").
			NewRangeReader(context.Background(), start, int64(chunk))
		defer rd.Close()

		end := math.Min(float64(start)+chunk, float64(rd.Size())-1)

		fmt.Println(fmt.Sprintf("bytes %v-%v/%v", start, int(end), rd.Size()))

		c.Stream(func(w io.Writer) bool {
			if int64(end) < (rd.Size() - 1) {
				c.DataFromReader(
					http.StatusOK,
					int64(chunk), "video/mp4",
					rd, map[string]string{
						"Accept-Ranges": "bytes",
						"Content-Range": fmt.Sprintf("bytes %v-%v/%v", start, int(end), rd.Size()),
					})
				c.SSEvent("s", "")
				return true
			}

			return false
		})
	})

	router.Run()
}
