package client

import (
	"context"

	"nhooyr.io/websocket"
)

type heartbeatService struct {
	ctx context.Context

	gatewayWebsocket *websocket.Conn
}

func NewHeartbeatService(ctx context.Context, gatewayWebsocket *websocket.Conn) (*heartbeatService, error) {
	return &heartbeatService{
		ctx:              ctx,
		gatewayWebsocket: gatewayWebsocket,
	}, nil
}

func (s *heartbeatService) Start() error {

	return nil
}
