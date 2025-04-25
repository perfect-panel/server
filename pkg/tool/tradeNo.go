package tool

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

func GenerateTradeNo() string {
	now := time.Now()
	formattedTime := now.Format("20060102150405") + strconv.Itoa(now.Nanosecond())
	numeric := [10]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	r := len(numeric)
	source := rand.NewSource(time.Now().UnixNano())
	random := rand.New(source)
	var code strings.Builder
	for i := 0; i < 4; i++ {
		_, _ = fmt.Fprintf(&code, "%d", numeric[random.Intn(r)])
	}
	formattedTime += code.String()
	return formattedTime
}
