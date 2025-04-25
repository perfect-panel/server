package device

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/pkg/errors"

	"github.com/gorilla/websocket"
)

func TestDevice(t *testing.T) {
	t.Skip("skip test")
	/*	deviceManager := NewDeviceManager(10, 3)

		deviceManager.OnDeviceOnline = func(userID int64, deviceID, session string) {
			fmt.Printf("✅ 设备 %s (用户 %d) 上线\n", deviceID, userID)
		}

		deviceManager.OnDeviceOffline = func(userID int64, deviceID, session string) {
			fmt.Printf("❌ 设备 %s (用户 %d) 下线\n", deviceID, userID)
		}

		deviceManager.OnDeviceKicked = func(userID int64, deviceID, session string, operator Operator) {
			fmt.Printf("⚠️ 设备 %s (用户 %d) 被踢下线\n", deviceID, userID)
		}
		deviceManager.OnMessage = func(userID int64, deviceID, session string, message string) {
			log.Printf("✅收到消息:  设备 %s (用户 %d) 内容: %s,sesion: %s\n", deviceID, userID, message, session)
		}
		engine := gin.Default()
		engine.GET("/ws/:userid/:device_number", func(c *gin.Context) {
			//根据Authorization获取session
			authorization := c.GetHeader("Authorization")
			userid, err := strconv.ParseInt(c.Param("userid"), 10, 64)
			if err != nil {
				t.Errorf("get user id err:%v", err)
				return
			}
			deviceNumber := c.Param("device_number")
			deviceManager.AddDevice(c, authorization, userid, deviceNumber, 3)
			return
		})
		go func() {
			err := http.ListenAndServe(":8081", engine)
			if err != nil {
				t.Fatalf("engine start failed: %v", err)
			}
		}()
	*/
	h := http.Header{}
	h.Add("Authorization", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJTZXNzaW9uSWQiOiIwMTk0Y2ZiNy1hYjY0LTdjYjMtODUzYi03ZGU5YTAzNWRlZTgiLCJVc2VySWQiOjI5LCJleHAiOjE3MzkyNTY1MDgsImlhdCI6MTczODY1MTcwOH0.BGKT5-hongJPZrA_yAb6cf6go5iDR8T9uu1ZxUg8HDw")

	mutex := sync.Mutex{}
	serverURL := fmt.Sprintf("ws://localhost:8080/v1/app/ws/%d/%s", 29, "15502502051") // 假设 userID 为 1001，设备ID 为 deviceA

	// 建立 WebSocket 连接
	conn, resp, err := websocket.DefaultDialer.Dial(serverURL, h)
	if err != nil {
		all, err := io.ReadAll(resp.Body)
		t.Fatalf("websocket dial failed: %v:%s", err, string(all))
	}
	// 启动一个 goroutine 来读取服务器消息
	go func() {
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				if errors.Is(err, net.ErrClosed) || strings.Contains(err.Error(), "use of closed network connection") {
					log.Println("连接已关闭")
					return
				}
				log.Printf("接收消息失败: %v", err)
				return
			}
			fmt.Printf("收到来自服务器的消息: %s\n", msg)
		}
	}()

	//发送心跳
	go func() {
		ticker := time.NewTicker(time.Second * 5)
		defer ticker.Stop()

		for range ticker.C {
			mutex.Lock()
			err := conn.WriteMessage(websocket.TextMessage, []byte("ping"))
			mutex.Unlock()

			if err != nil {
				if strings.Contains(err.Error(), "use of closed network connection") {
					log.Println("连接已关闭")
					return
				}
				t.Errorf("websocket 写入失败: %v", err)
				return
			}
		}
	}()

	updateSubscribe, _ := json.Marshal(map[string]interface{}{
		"method": "test_method",
	})

	//发送一条消息
	mutex.Lock()
	err = conn.WriteMessage(websocket.TextMessage, updateSubscribe)
	mutex.Unlock()
	if err != nil {
		t.Errorf("websocket write failed: %v", err)
	}

	time.Sleep(time.Second * 20)
	conn.Close()
	time.Sleep(time.Second * 5)

}
