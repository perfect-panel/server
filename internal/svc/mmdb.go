package svc

import (
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/oschwald/geoip2-golang"
	"github.com/perfect-panel/server/pkg/logger"
)

const GeoIPDBURL = "https://raw.githubusercontent.com/adysec/IP_database/main/geolite/GeoLite2-City.mmdb"

type IPLocation struct {
	Path string
	DB   *geoip2.Reader
}

func NewIPLocation(path string) (*IPLocation, error) {

	// 检查文件是否存在
	if _, err := os.Stat(path); os.IsNotExist(err) {
		logger.Infof("[GeoIP] Database not found, downloading from %s", GeoIPDBURL)
		// 文件不存在，下载数据库
		err := DownloadGeoIPDatabase(GeoIPDBURL, path)
		if err != nil {
			logger.Errorf("[GeoIP] Failed to download database: %v", err.Error())
			return nil, err
		}
		logger.Infof("[GeoIP] Database downloaded successfully")
	}

	db, err := geoip2.Open(path)
	if err != nil {
		return nil, err
	}
	return &IPLocation{
		Path: path,
		DB:   db,
	}, nil
}

func (ipLoc *IPLocation) Close() error {
	return ipLoc.DB.Close()
}

func DownloadGeoIPDatabase(url, path string) error {

	// 创建路径, 确保目录存在
	err := os.MkdirAll(filepath.Dir(path), 0755)
	if err != nil {
		logger.Errorf("[GeoIP] Failed to create directory: %v", err.Error())
		return err
	}

	// 创建文件
	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer out.Close()

	// 请求远程文件
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 保存文件
	_, err = io.Copy(out, resp.Body)
	return err
}
