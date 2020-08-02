package upload

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"syscall"
	"time"
	"unsafe"

	"lft/cache"
	"lft/global"

	// 打包用
	_ "lft/statik"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rakyll/statik/fs"
)

var localIP string

func init() {
	localIP = findIP()
}

// StartServer 启动服务
func StartServer(port int) error {
	uploadServer := &http.Server{
		Addr:         fmt.Sprintf("%s%d", "localhost:", port),
		Handler:      router(),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	log.Println("请使用浏览器访问：", "http://localhost:8089/static")
	return uploadServer.ListenAndServe()
}

func router() http.Handler {
	statikFS, err := fs.New()
	if err != nil {
		log.Fatalf("server run error: %v", err)
	}

	e := gin.New()
	e.Use(gin.Recovery())
	e.Use(gin.Logger())
	e.StaticFS("static", statikFS)
	e.POST("/rest/upload", upload)
	e.GET("/rest/list", list)
	e.POST("/rest/delete", delete)

	return e
}

func upload(c *gin.Context) {
	json := make(map[string]string)
	c.ShouldBindJSON(&json)

	path, ok := json["path"]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"err": "path not found",
		})
		return
	}

	fileInfo, err := os.Stat(path)
	if os.IsNotExist(err) {
		c.JSON(http.StatusInternalServerError, gin.H{
			"err": "file not exist",
		})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"err": fmt.Sprintf("get file info error: %v", err),
		})
		return
	}

	if fileInfo.IsDir() {
		c.JSON(http.StatusInternalServerError, gin.H{
			"err": "not support dir",
		})
		return
	}

	key, err := uuid.NewUUID()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"err": fmt.Sprintf("uuid error: %v", err),
		})
		return
	}
	global.FileMap.Add(key.String(), path)
	// 显示当前缓存中所有路径
	c.JSON(http.StatusOK, gin.H{
		"downloadUrls": convertStr(global.FileMap.List()),
	})
}

func list(c *gin.Context) {
	// 显示当前缓存中所有路径
	c.JSON(http.StatusOK, gin.H{
		"downloadUrls": convertStr(global.FileMap.List()),
	})
}

func delete(c *gin.Context) {
	json := make(map[string]string)
	c.ShouldBindJSON(&json)

	id, ok := json["id"]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"err": "param id not found",
		})
		return
	}

	global.FileMap.Remove(id)

	// 显示当前缓存中所有路径
	c.JSON(http.StatusOK, gin.H{
		"downloadUrls": convertStr(global.FileMap.List()),
	})
}

// Pair 是 DTO 数据格式
type Pair struct {
	URL  string `json:"url"`
	Path string `json:"path"`
}

func convertStr(list []*cache.Entry) []Pair {
	strList := make([]Pair, 0)
	for _, val := range list {
		strList = append(strList, Pair{
			URL:  fmt.Sprintf("http://%s:%d/rest/download/%s", findIP(), global.DownloadPort, val.Key),
			Path: val.Value.(string),
		})
	}
	return strList
}

func findIP() string {
	indexs := make([]int, 0)
	netInterfaces, err := net.Interfaces()
	if err != nil {
		return "127.0.0.1"
	}

	for i := 0; i < len(netInterfaces); i++ {
		if (netInterfaces[i].Flags & net.FlagUp) == net.FlagUp {
			addrs, _ := netInterfaces[i].Addrs()
			for _, address := range addrs {
				if ipnet, ok := address.(*net.IPNet); ok &&
					!ipnet.IP.IsLoopback() &&
					ipnet.IP.To4() != nil {
					indexs = append(indexs, netInterfaces[i].Index)
					break
				}
			}
		}
	}

	if len(indexs) > 0 {
		aList, err := getAdapterList()
		if err != nil {
			return "127.0.0.1"
		}

		for ai := aList; ai != nil; ai = ai.Next {
			index := ai.Index
			for _, i := range indexs {
				if int(index) == i {
					ipl := &ai.IpAddressList
					gwl := &ai.GatewayList
					for ; ipl != nil; ipl = ipl.Next {
						if string(removeZore(gwl.IpAddress.String)) != "0.0.0.0" {
							return string(removeZore(ipl.IpAddress.String))
						}
					}
				}
			}
		}
	}
	return "127.0.0.1"
}

func getAdapterList() (*syscall.IpAdapterInfo, error) {
	b := make([]byte, 1000)
	l := uint32(len(b))
	a := (*syscall.IpAdapterInfo)(unsafe.Pointer(&b[0]))
	err := syscall.GetAdaptersInfo(a, &l)
	if err == syscall.ERROR_BUFFER_OVERFLOW {
		b = make([]byte, l)
		a = (*syscall.IpAdapterInfo)(unsafe.Pointer(&b[0]))
		err = syscall.GetAdaptersInfo(a, &l)
	}
	if err != nil {
		return nil, os.NewSyscallError("GetAdaptersInfo", err)
	}
	return a, nil
}

func removeZore(bytes [16]byte) []byte {
	ret := make([]byte, 0)
	for _, b := range bytes {
		if b == 0 {
			break
		}
		ret = append(ret, b)
	}
	return ret
}
