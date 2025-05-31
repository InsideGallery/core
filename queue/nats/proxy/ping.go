package proxy

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"time"
)

type Ping struct {
	client    NATSRequester
	pingEvery time.Duration
}

func NewPing(client NATSRequester, pingEvery time.Duration) *Ping {
	return &Ping{
		client:    client,
		pingEvery: pingEvery,
	}
}

func (p *Ping) Ping(srv *Server, instanceID string) error {
	resp, err := p.client.Requester(instanceID, []byte(PingMsg), time.Second)
	if err != nil {
		return errors.Join(err, srv.balancer.DestroyInstance(instanceID))
	}

	if !strings.EqualFold(string(resp), PongMsg) {
		return errors.Join(ErrWrongPongResponse, srv.balancer.DestroyInstance(instanceID))
	}

	return nil
}

func (p *Ping) Service(ctx context.Context, srv *Server) {
	ticker := time.NewTicker(p.pingEvery)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			for _, iid := range srv.GetInstances() {
				err := p.Ping(srv, iid)
				if err != nil {
					slog.Default().Error("error ping instance, remove from instances", "iid", iid, "err", err)
				}
			}
		}
	}
}
