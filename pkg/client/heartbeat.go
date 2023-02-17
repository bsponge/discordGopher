package client

import (
	"context"
	"time"

	"github.com/bsponge/discordGopher/pkg/log"

	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

type heartbeatService struct {
	ctx    context.Context
	closed chan struct{}

	client *Client
}

type heartbeat struct {
	Op int `json:"op"`
	D  int `json:"d"`
}

type SequenceGetter interface {
	GetSequence() int
}

func NewHeartbeatService(ctx context.Context, client *Client) *heartbeatService {
	return &heartbeatService{
		ctx:    ctx,
		closed: make(chan struct{}),
		client: client,
	}
}

func (s *heartbeatService) SendHeartbeat(gatewayWebsocket *websocket.Conn) error {
	hb := &heartbeat{
		Op: 1,
		D:  s.client.GetSequence(),
	}

	log.Logger().Trace("Sending heartbeat")

	err := wsjson.Write(s.ctx, gatewayWebsocket, hb)
	if err != nil {
		return err
	}

	return nil
}

func (s *heartbeatService) Start(gatewayWebsocket *websocket.Conn, interval int) error {
	timer := time.NewTicker(time.Duration(interval) * time.Millisecond)

	err := s.SendHeartbeat(gatewayWebsocket)
	if err != nil {
		return err
	}

	err = s.client.Identify()
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case <-s.ctx.Done():
				close(s.closed)
				return
			case <-timer.C:
			}

			err := s.SendHeartbeat(gatewayWebsocket)
			if err != nil {
				log.Logger().WithError(err).Error("Could not send heartbeat")
			}
		}
	}()

	return nil
}

func (s *heartbeatService) Stop() {
	<-s.closed
}
