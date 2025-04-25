package device

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"go.uber.org/zap"

	"github.com/gorilla/websocket"
)

type Operator int

const (
	MaxDevices Operator = iota
	Admin
	SubscribeUpdate = "subscribe_update"
)

// Device represents a device structure
type Device struct {
	Session      string
	DeviceID     string
	Conn         *websocket.Conn
	CreatedAt    time.Time
	LastPingTime time.Time
}

// WebSocket upgrader
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// DeviceManager manages devices
type DeviceManager struct {
	userDevices      sync.Map // userID -> []*Device
	totalOnline      int32    // total online devices
	userMutexes      sync.Map // userID level locks
	heartbeatTimeout int      // heartbeat timeout (seconds)
	checkInterval    int      // heartbeat check interval (seconds)

	// event callbacks
	OnDeviceOnline  func(userID int64, deviceID, session string)
	OnDeviceOffline func(userID int64, deviceID, session string, createAt time.Time)
	OnDeviceKicked  func(userID int64, deviceID, session string, operator Operator)
	OnMessage       func(userID int64, deviceID, session string, message string)
}

// Get user-level mutex
func (dm *DeviceManager) getUserMutex(userID int64) *sync.Mutex {
	mu, _ := dm.userMutexes.LoadOrStore(userID, &sync.Mutex{})
	return mu.(*sync.Mutex)
}

// Listen to WebSocket data
func (dm *DeviceManager) listenToDevice(userID int64, device *Device) {
	defer func() {
		dm.removeDevice(userID, device.DeviceID) // remove device when disconnected
	}()

	for {
		_, msg, err := device.Conn.ReadMessage()
		if err != nil {
			zap.S().Infof("Device %s (User %d) disconnected: %v", device.DeviceID, userID, err)
			break
		}

		message := string(msg)
		if message == "ping" || message == "heartbeat" {
			dm.UpdateHeartbeat(userID, device.DeviceID)
			continue
		}

		// Trigger message callback
		if dm.OnMessage != nil {
			go dm.OnMessage(userID, device.DeviceID, device.Session, message)
		}
	}
}

// UpdateHeartbeat updates device heartbeat
func (dm *DeviceManager) UpdateHeartbeat(userID int64, deviceID string) {
	mu := dm.getUserMutex(userID)
	mu.Lock()
	defer mu.Unlock()

	if val, ok := dm.userDevices.Load(userID); ok {
		devices := val.([]*Device)
		for _, d := range devices {
			if d.DeviceID == deviceID {
				d.LastPingTime = time.Now()
				if err := d.Conn.WriteMessage(websocket.TextMessage, []byte("ping")); err != nil {
					zap.S().Infof("✅ Heartbeat updated: Device %s (User %d) err: %s", deviceID, userID, err.Error())
				}
				break
			}
		}
	}
}

// AddDevice **Add: Device connects WebSocket and is added to the manager**
func (dm *DeviceManager) AddDevice(w http.ResponseWriter, r *http.Request, session string, userID int64, deviceID string, maxDevices int) {
	// **Upgrade WebSocket connection**
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		zap.S().Infof("WebSocket upgrade failed: %v", err)
		return
	}

	mu := dm.getUserMutex(userID)
	mu.Lock()
	defer mu.Unlock()

	newDevice := &Device{
		Session:      session,
		DeviceID:     deviceID,
		Conn:         conn,
		CreatedAt:    time.Now(),
		LastPingTime: time.Now(),
	}

	//不限制设备数量
	if maxDevices < 1 {
		maxDevices = 99
	}

	// Get user's device list
	var restConnection bool
	var devices []*Device
	if val, ok := dm.userDevices.Load(userID); ok {
		devices = val.([]*Device)
		var tempDevice []*Device
		for _, d := range devices {
			if d.DeviceID == deviceID {
				restConnection = true
			} else {
				tempDevice = append(tempDevice, d)
			}
		}
		devices = tempDevice
	}

	// **If exceeding the limit, kick out the earliest device**
	if !restConnection && len(devices) >= maxDevices {
		oldestDevice := devices[0]
		devices = devices[1:]

		if dm.OnDeviceKicked != nil {
			done := make(chan struct{})
			go func() {
				defer close(done)
				dm.OnDeviceKicked(userID, oldestDevice.DeviceID, oldestDevice.Session, MaxDevices)
			}()
			<-done // block and wait for callback to complete
		}
		oldestDevice.Conn.Close()
		atomic.AddInt32(&dm.totalOnline, -1)
	}

	// Add new device
	devices = append(devices, newDevice)
	dm.userDevices.Store(userID, devices)
	atomic.AddInt32(&dm.totalOnline, 1)

	// Trigger online event
	if dm.OnDeviceOnline != nil {
		go dm.OnDeviceOnline(userID, deviceID, session)
	}

	// Start listening
	go dm.listenToDevice(userID, newDevice)
}

// removeDevice removes a device
func (dm *DeviceManager) removeDevice(userID int64, deviceID string) {
	mu := dm.getUserMutex(userID)
	mu.Lock()
	defer mu.Unlock()

	if val, ok := dm.userDevices.Load(userID); ok {
		devices := val.([]*Device)
		for i, d := range devices {
			if d.DeviceID == deviceID {
				devices = append(devices[:i], devices[i+1:]...)
				d.Conn.Close()
				atomic.AddInt32(&dm.totalOnline, -1)

				if dm.OnDeviceOffline != nil {
					go dm.OnDeviceOffline(userID, deviceID, d.Session, d.CreatedAt)
				}
				break
			}
		}

		if len(devices) == 0 {
			dm.userDevices.Delete(userID)
		} else {
			dm.userDevices.Store(userID, devices)
		}
	}
}

// KickDevice kicks a device (supports individual device or entire user)
func (dm *DeviceManager) KickDevice(userID int64, deviceID string) {
	mu := dm.getUserMutex(userID)
	mu.Lock()
	defer mu.Unlock()

	// Get user's device list
	val, ok := dm.userDevices.Load(userID)
	if !ok {
		zap.S().Infof("⚠️ User %d has no online devices, unable to kick out", userID)
		return
	}

	devices := val.([]*Device)
	var activeDevices []*Device

	for _, d := range devices {
		if deviceID == "" || d.DeviceID == deviceID {
			// Trigger kick event callback
			if dm.OnDeviceKicked != nil {
				done := make(chan struct{})
				go func() {
					defer close(done)
					dm.OnDeviceKicked(userID, d.DeviceID, d.Session, Admin)
				}()
				<-done // block and wait for callback to complete
			}
			// Close WebSocket connection
			d.Conn.Close()
			atomic.AddInt32(&dm.totalOnline, -1)
			zap.S().Infof("❌ Device %s (User %d) kicked out", d.DeviceID, userID)
		} else {
			activeDevices = append(activeDevices, d)
		}
	}

	// Update user's device mapping
	if len(activeDevices) == 0 {
		dm.userDevices.Delete(userID)
	} else {
		dm.userDevices.Store(userID, activeDevices)
	}
}

// StartHeartbeatCheck periodically checks for heartbeat timeout devices
func (dm *DeviceManager) StartHeartbeatCheck() {
	ticker := time.NewTicker(time.Duration(dm.checkInterval) * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now()

		dm.userDevices.Range(func(userID, val interface{}) bool {
			uid := userID.(int64)
			devices := val.([]*Device)

			mu := dm.getUserMutex(uid)
			mu.Lock()
			defer mu.Unlock()

			var activeDevices []*Device
			for _, d := range devices {
				if now.Sub(d.LastPingTime) > time.Duration(dm.heartbeatTimeout)*time.Second {
					zap.S().Infof("⚠️ Device %s (User %d) heartbeat timeout, removed", d.DeviceID, uid)
					d.Conn.Close()
					atomic.AddInt32(&dm.totalOnline, -1)

					if dm.OnDeviceOffline != nil {
						go dm.OnDeviceOffline(uid, d.DeviceID, d.Session, d.CreatedAt)
					}
				} else {
					activeDevices = append(activeDevices, d)
				}
			}

			if len(activeDevices) == 0 {
				dm.userDevices.Delete(uid)
			} else {
				dm.userDevices.Store(uid, activeDevices)
			}
			return true
		})
		//zap.S().Infof("Total online devices: %d\n", dm.totalOnline)
	}
}

// NewDeviceManager creates a new device manager
func NewDeviceManager(heartbeatTimeout, checkInterval int) *DeviceManager {
	dm := &DeviceManager{
		heartbeatTimeout: heartbeatTimeout,
		checkInterval:    checkInterval,
	}
	go dm.StartHeartbeatCheck()
	return dm
}

// SendToDevice sends a message to a specific device
func (dm *DeviceManager) SendToDevice(userID int64, deviceID string, message string) error {
	if val, ok := dm.userDevices.Load(userID); ok {
		devices := val.([]*Device)
		if deviceID == "" {
			for _, d := range devices {
				err := d.Conn.WriteMessage(websocket.TextMessage, []byte(message))
				if err != nil {
					return err
				}
				continue
			}
		} else {
			for _, d := range devices {
				if d.DeviceID == deviceID {
					return d.Conn.WriteMessage(websocket.TextMessage, []byte(message))
				}
			}
		}

	}
	return fmt.Errorf("device %s (User %d) is offline", deviceID, userID)
}

// Broadcast sends a message to all devices
func (dm *DeviceManager) Broadcast(message string) {
	go func(message string) {
		dm.userDevices.Range(func(_, val interface{}) bool {
			devices := val.([]*Device)
			for _, d := range devices {
				_ = d.Conn.WriteMessage(websocket.TextMessage, []byte(message))
			}
			return true
		})
	}(message)

}

// Gracefully shut down all WebSocket connections
func (dm *DeviceManager) Shutdown(ctx context.Context) {
	<-ctx.Done()
	zap.S().Infof("🔴 Shutting down all WebSocket connections...")

	dm.userDevices.Range(func(userID, val interface{}) bool {
		uid := userID.(int64)
		devices := val.([]*Device)

		for _, d := range devices {
			d.Conn.Close()
			zap.S().Infof("✅ Closed device %s (User %d)", d.DeviceID, uid)
		}
		dm.userDevices.Delete(uid)
		return true
	})
}
