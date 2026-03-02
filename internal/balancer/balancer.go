package balancer

import (
	"lb/internal/backend"
)

type Balancer interface {
	Select() *backend.Backend
}

type SmoothWRR struct {
	pool *backend.ServerPool
}

func NewSmoothWRR(pool *backend.ServerPool) Balancer {
	return &SmoothWRR{pool: pool}
}

func (b *SmoothWRR) Select() *backend.Backend {
	return b.pool.GetNextPeer()
}
