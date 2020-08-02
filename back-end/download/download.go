package download

import (
	"fmt"
	"lft/global"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// StartServer 启动服务
func StartServer(port int) error {
	downloadServer := &http.Server{
		Addr:         fmt.Sprintf("%s%d", ":", port),
		Handler:      router(),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	return downloadServer.ListenAndServe()
}

func router() http.Handler {
	e := gin.New()
	e.Use(gin.Recovery())
	e.Use(gin.Logger())
	e.GET("/rest/download/:uuid", download)

	return e
}

func download(c *gin.Context) {
	id := c.Param("uuid")
	filePath, ok := global.FileMap.Get(id)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{
			"err": "file path not found",
		})
		return
	}

	file, err := os.OpenFile(filePath.(string), os.O_RDONLY, 0666)
	defer file.Close()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"err": fmt.Sprintf("open file error: %v", err),
		})
		return
	}

	c.Writer.Header().Add("Content-Type", "application/octet-stream")
	c.Writer.Header().Add("Content-Disposition", fmt.Sprintf("attachment;filename=%s", filepath.Base(filePath.(string))))
	fileInfo, err := os.Stat(filePath.(string))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"err": fmt.Sprintf("open file error: %v", err),
		})
		return
	}

	fileSize := fileInfo.Size()
	c.Writer.Header().Add("Last-Modified", fileInfo.ModTime().Format(time.RFC1123))
	c.Writer.Header().Add("ETag", fileInfo.ModTime().Format(time.RFC1123))
	c.Writer.Header().Add("Accept-Ranges", "bytes")

	rangeValue := c.GetHeader("Range")
	fmt.Println(rangeValue)
	if rangeValue == "" {
		c.File(filePath.(string))
		return
	}
	ranges := parseRange(rangeValue, fileSize)
	start := ranges[0].start
	buff := make([]byte, 1024*1024*4)
	count, _ := file.ReadAt(buff, start)
	c.Writer.Header().Add("Content-Length", strconv.Itoa(count))
	c.Writer.Header().Add(
		"Content-Range",
		fmt.Sprintf("bytes %d-%d/%d", start, start+int64(count)-1, fileSize))
	c.Status(http.StatusPartialContent)
	c.Writer.Write(buff[0:count])
}

type rangee struct {
	start int64
	end   int64
}

func parseRange(rangeStr string, fileSize int64) []rangee {
	ranges := make([]rangee, 0)
	rangeStr = strings.ReplaceAll(rangeStr, "bytes=", "")
	rangeStrs := strings.Split(rangeStr, ",")
	for _, str := range rangeStrs {
		str = strings.Trim(str, " ")
		if strings.HasPrefix(str, "-") {
			endSize, _ := strconv.ParseInt(str, 10, 64)
			ranges = append(ranges, rangee{
				start: fileSize + endSize,
				end:   fileSize - 1,
			})
		} else if strings.HasSuffix(str, "-") {
			startIndex, _ := strconv.ParseInt(strings.TrimSuffix(str, "-"), 10, 64)
			ranges = append(ranges, rangee{
				start: startIndex,
				end:   fileSize - 1,
			})
		} else {
			r := strings.Split(str, "-")
			start, _ := strconv.ParseInt(r[0], 10, 64)
			end, _ := strconv.ParseInt(r[1], 10, 64)
			ranges = append(ranges, rangee{
				start: start,
				end:   end - 1,
			})
		}
	}
	return ranges
}
