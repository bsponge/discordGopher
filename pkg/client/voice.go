package client

import (
	"context"

	"github.com/bsponge/discordGopher/pkg/object"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

type voiceClient struct {
	ctx context.Context

	client         *Client
	voiceWebsocket *websocket.Conn

	voiceServerUpdateCh chan object.VoiceServerUpdate
	voiceStateCh        chan object.VoiceState
}

func NewVoiceClient(ctx context.Context, client *Client) *voiceClient {
	return &voiceClient{
		ctx:    ctx,
		client: client,
	}
}

func (c *voiceClient) ConnectToVoiceChannel(guildID string, channelID string, selfMute bool, selfDeaf bool) error {
	voiceEvent := object.Event[object.VoiceState]{
		Op: 4,
		D: object.VoiceState{
			GuildID:   &guildID,
			ChannelID: &channelID,
			SelfMute:  &selfMute,
			SelfDeaf:  &selfDeaf,
		},
	}

	c.voiceServerUpdateCh = make(chan object.VoiceServerUpdate)
	c.voiceStateCh = make(chan object.VoiceState)

	err := wsjson.Write(c.ctx, c.client.gatewayWebsocket, voiceEvent)
	if err != nil {
		return err
	}

	var voiceServerUpdate object.VoiceServerUpdate
	var voiceState object.VoiceState

	select {
	case voiceServerUpdate = <-c.voiceServerUpdateCh:
	case voiceState = <-c.voiceStateCh:
	}

	select {
	case voiceServerUpdate = <-c.voiceServerUpdateCh:
	case voiceState = <-c.voiceStateCh:
	}

	ws, _, err := websocket.Dial(c.ctx, voiceServerUpdate.Endpoint, nil)
	if err != nil {
		return err
	}

	c.voiceWebsocket = ws

	identifyEvent := object.Event[object.VoiceIdentify]{
		Op: 0,
		D: object.VoiceIdentify{
			ServerID:  voiceServerUpdate.GuildID,
			UserID:    voiceState.UserID,
			SessionID: voiceState.SessionID,
			Token:     voiceServerUpdate.Token,
		},
	}
	_ = identifyEvent // TODO identify

	return nil
}

func (c *voiceClient) GetVoiceServerUpdateCh() chan object.VoiceServerUpdate {
	return c.voiceServerUpdateCh
}

func (c *voiceClient) GetVoiceStateCh() chan object.VoiceState {
	return c.voiceStateCh
}
