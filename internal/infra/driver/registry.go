package driver

import (
	"fmt"
	"sync"
)

// Registry 驱动注册中心，按 DriverType 索引具体实现
type Registry struct {
	mu      sync.RWMutex
	drivers map[DriverType]Driver
}

// defaultRegistry 全局默认注册中心
var defaultRegistry = &Registry{
	drivers: make(map[DriverType]Driver),
}

// Register 注册一个驱动实现到全局注册中心
func Register(d Driver) {
	defaultRegistry.Register(d)
}

// Get 从全局注册中心获取驱动
func Get(t DriverType) (Driver, error) {
	return defaultRegistry.Get(t)
}

// MustGet 获取驱动，找不到则 panic
func MustGet(t DriverType) Driver {
	d, err := Get(t)
	if err != nil {
		panic(err)
	}
	return d
}

// Register 把驱动放入注册中心
func (r *Registry) Register(d Driver) {
	if d == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.drivers[d.Type()] = d
}

// Get 按类型获取驱动
func (r *Registry) Get(t DriverType) (Driver, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if d, ok := r.drivers[t]; ok {
		return d, nil
	}
	return nil, fmt.Errorf("driver: %q not registered", t)
}