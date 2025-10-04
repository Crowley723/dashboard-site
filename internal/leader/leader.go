package leader

import (
	"homelab-dashboard/internal/metrics"
	"homelab-dashboard/internal/middlewares"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

type Election struct {
	Redis      *redis.Client
	InstanceID string        // Unique identifier (hostname, pod name, etc.)
	TTL        time.Duration // heartbeat timeout
	isLeader   bool
	mu         sync.RWMutex
}

func (e *Election) IsLeader() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.isLeader
}

func (e *Election) campaign(ctx middlewares.AppContext) {
	ok, err := e.Redis.SetNX(ctx, "leader", e.InstanceID, e.TTL).Result()
	if err != nil {
		return
	}

	e.mu.Lock()
	wasLeader := e.isLeader
	e.isLeader = ok
	e.mu.Unlock()

	if ok && !wasLeader {
		ctx.Logger.Info("became leader", "instance", e.InstanceID)
		metrics.IsLeader.Set(1)
	} else if !ok && wasLeader {
		ctx.Logger.Info("lost leadership", "instance", e.InstanceID)
		metrics.IsLeader.Set(0)
	}

	if ok {
		e.Redis.Expire(ctx, "leader", e.TTL)
	}
}

func (e *Election) Start(ctx middlewares.AppContext) {
	ticker := time.NewTicker(e.TTL / 3)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			e.resign(ctx)
			return
		case <-ticker.C:
			e.campaign(ctx)
		}
	}
}

func (e *Election) resign(ctx middlewares.AppContext) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if !e.isLeader {
		return
	}

	script := `
        if redis.call("get", KEYS[1]) == ARGV[1] then
            return redis.call("del", KEYS[1])
        end
        return 0
    `
	redis.NewScript(script).Run(ctx, e.Redis, []string{"leader"}, e.InstanceID)
	e.isLeader = false
}
