package ip

import (
	"testing"
	"time"
)

func TestGetIPv4(t *testing.T) {
	t.Skip("skip TestGetIPv4")
	iPv4, err := GetIP("baidu.com")
	if err != nil {
		t.Fatal(err)
	}

	t.Log(iPv4)
}

func TestGetRegionByIp(t *testing.T) {
	t.Skip("skip TestGetRegionByIp")
	ips, err := GetIP("122.14.229.128")
	if err != nil {
		t.Fatal(err)
	}

	for _, ip := range ips {
		t.Log(ip)
		resp, err := GetRegionByIp(ip)
		if err != nil {
			t.Fatalf("ip: %s,err: %v", ip, err)
		}
		t.Logf("country: %s,City: %s,latitude:%s, longitude:%s", resp.Country, resp.City, resp.Latitude, resp.Longitude)
	}
	time.Sleep(3 * time.Second)
}
