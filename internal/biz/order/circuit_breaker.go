package order

import (
	"sync"
	"time"
)

// circuitBreaker 是一个轻量级熔断器，保护本服务不被持续不可用的下游（订单中台）拖垮。
//
// 状态机：
//   - closed：正常放行；连续失败达到 failThreshold 后切到 open。
//   - open：快速失败，不再发起下游调用；经过 cooldown 后进入 half-open。
//   - half-open：允许一次探测调用，成功则恢复 closed，失败则回到 open。
type circuitBreaker struct {
	mu            sync.Mutex
	state         cbState
	consecutiveFails int
	failThreshold int
	cooldown      time.Duration
	openedAt      time.Time
}

type cbState int

const (
	cbClosed cbState = iota
	cbOpen
	cbHalfOpen
)

// newCircuitBreaker 创建熔断器。failThreshold 为连续失败阈值，cooldown 为熔断冷却时间。
func newCircuitBreaker(failThreshold int, cooldown time.Duration) *circuitBreaker {
	return &circuitBreaker{
		state:        cbClosed,
		failThreshold: failThreshold,
		cooldown:      cooldown,
	}
}

// allow 返回当前是否允许发起下游调用。open 状态且未过冷却期则拒绝。
func (cb *circuitBreaker) allow() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	switch cb.state {
	case cbClosed, cbHalfOpen:
		return true
	case cbOpen:
		if time.Since(cb.openedAt) >= cb.cooldown {
			cb.state = cbHalfOpen
			return true
		}
		return false
	default:
		return true
	}
}

// recordSuccess 记录一次成功，重置失败计数并回到 closed。
func (cb *circuitBreaker) recordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.consecutiveFails = 0
	cb.state = cbClosed
}

// recordFailure 记录一次失败，达到阈值则熔断。
func (cb *circuitBreaker) recordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.consecutiveFails++
	if cb.state == cbHalfOpen || cb.consecutiveFails >= cb.failThreshold {
		cb.state = cbOpen
		cb.openedAt = time.Now()
	}
}
