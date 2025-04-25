package snowflake

import (
	"errors"
	"net"

	sf "github.com/GUAIK-ORG/go-snowflake/snowflake"
)

var snowflake *sf.Snowflake

func init() {
	localIp3, err := GetLocalIp()
	if err != nil {
		panic(err)
	}
	dataCenterID := int64(localIp3 >> 4)
	workerID := int64(localIp3 & 0xf)
	snowflake, err = sf.NewSnowflake(dataCenterID, workerID)
	if err != nil {
		panic(err)
	}
}

func GetID() int64 {
	return snowflake.NextVal()
}

func GetTimestamp(id int64) int64 {
	return sf.GetTimestamp(id)
}

func GetGenTimestamp(id int64) int64 {
	return sf.GetGenTimestamp(id)
}

func GetLocalIp() (byte, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return 0, err
	}

	for _, i := range ifaces {
		addrs, errRet := i.Addrs()
		if errRet != nil {
			continue
		}

		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				if v.IP.IsGlobalUnicast() {
					ip = v.IP.To4()
					if ip != nil {
						return ip[3], nil
					}
				}
			}
		}
	}

	return 0, errors.New("no validate ifaces to IPV4")
}
