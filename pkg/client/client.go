package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"runtime"
	"sync"

	"github.com/bsponge/discordGopher/pkg/config"
	"github.com/bsponge/discordGopher/pkg/errors"
	"github.com/bsponge/discordGopher/pkg/log"
	"github.com/bsponge/discordGopher/pkg/object"

	"github.com/valyala/fastjson"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

const (
	apiVersionKey = "v"
	encodingKey   = "encoding"

	apiVersionValue = "10"
	encodingValue   = "json"

	tokenURL    = "https://discord.com/api/oauth2/token"
	apiEndpoint = "https://discord.com/api/v10"
)

type Client struct {
	ctx    context.Context
	cancel context.CancelFunc

	mtx sync.Mutex

	cfg *config.Config

	sequence int

	gatewayWebsocket *websocket.Conn
	resumeGatewayURL *url.URL

	hbService *heartbeatService
}

func NewClient(ctx context.Context) (*Client, error) {
	cfg, err := config.LoadConfig("")
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(ctx)

	client := &Client{
		ctx:    ctx,
		cancel: cancel,
		cfg:    cfg,
	}

	hbService := NewHeartbeatService(ctx, client)
	client.hbService = hbService

	return client, nil
}

func (c *Client) Start() error {
	log.Logger().Trace("Starting the client...")

	gatewayURL, err := c.getGatewayURL()
	if err != nil {
		return err
	}

	log.Logger().Infof("Gateway URL: %s", gatewayURL)

	ws, _, err := websocket.Dial(c.ctx, gatewayURL, nil)
	if err != nil {
		return err
	}

	c.gatewayWebsocket = ws

	go c.poolMessages()

	log.Logger().Trace("The client started successfully")

	return nil
}

func (c *Client) GetSequence() int {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	return c.sequence
}

func (c *Client) setSequence(sequence int) {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	c.sequence = sequence
}

func (c *Client) poolMessages() {
	for {
		_, body, err := c.gatewayWebsocket.Read(c.ctx)
		unwraped := errors.BaseError(err)
		if unwraped == context.Canceled {
			return
		}
		if err != nil {
			log.Logger().WithError(err).Error("Could not read message from gateway wss")
			continue
		}

		log.Logger().Trace(string(body))

		resp, err := fastjson.ParseBytes(body)
		if err != nil {
			log.Logger().WithError(err).Error("Could not parse json received from gateway wss")
			return
		}

		c.setSequence(resp.GetInt("s"))
		op := resp.GetInt("op")

		switch op {
		case 0: // Dispatch (most Gateway events which represent actions taking place in a guild)
			dispatch := string(resp.GetStringBytes("t"))
			var d []byte
			d = resp.GetObject("d").MarshalTo(d)
			c.handleDispatch(object.Dispatch(dispatch), d)
		case 1: // Extra heartbeat
			err := c.hbService.SendHeartbeat(c.gatewayWebsocket)
			if err != nil {
				log.Logger().WithError(err).Error("Could not send heartbeat after receiving op code 1")
			}
		case 9: // Invalid session
		case 10: // Hello
			heartbeatInterval := resp.Get("d").GetInt("heartbeat_interval")
			err := c.hbService.Start(c.gatewayWebsocket, heartbeatInterval)
			if err != nil {
				log.Logger().WithError(err).Error("Could not send first heartbeat")
				return
			}
		case 11: // Heartbeat ACK
			log.Logger().Trace("Received heartbeat ACK")
		default:
			log.Logger().Trace("Unknown op code")
		}
	}
}

func (c *Client) handleDispatch(dispatch object.Dispatch, payload []byte) error {
	switch dispatch {
	case object.ReadyType:
		var ready object.Ready
		err := json.Unmarshal(payload, &ready)
		if err != nil {
			return err
		}

		c.setResumeGatewayURL(ready.ResumeGatewayURL)
	case object.GuildCreate:
		var guild object.Guild
		err := json.Unmarshal(payload, &guild)
		if err != nil {
			return err
		}

		log.Logger().Info("Unmarshaled GUILD ", guild)
	default:
		log.Logger().Warn("Received unknown dispatch")
	}

	return nil
}

func (c *Client) setResumeGatewayURL(gatewayURL string) error {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	parsedURL, err := url.Parse(gatewayURL)
	if err != nil {
		return err
	}

	c.resumeGatewayURL = parsedURL

	return nil
}

func (c *Client) Identify() error {
	identify := object.Identify{
		Token: c.cfg.Token,
		Properties: map[string]any{
			"os":      runtime.GOOS,
			"browser": "disco",
			"device":  "disco",
		},
		Compress: false,
		Intents:  513, // TODO: what should go there?
	}

	e := object.Event[object.Identify]{
		Op: 2,
		D:  identify,
	}

	err := wsjson.Write(c.ctx, c.gatewayWebsocket, &e)
	if err != nil {
		return err
	}

	return nil
}

// getGateway gets gateway WSS URL which is used to listen for discord server events.
func (c *Client) getGatewayURL() (string, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/gateway/bot", apiEndpoint), nil)
	if err != nil {
		return "", err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bot %s", c.cfg.Token))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	gatewayURL := fastjson.GetString(body, "url")
	if gatewayURL == "" {
		return "", fmt.Errorf("could not obtain gateway url from received response")
	}

	url, err := url.Parse(gatewayURL)
	if err != nil {
		return "", err
	}

	values := url.Query()

	values.Add(apiVersionKey, apiVersionValue)
	values.Add(encodingKey, encodingValue)

	url.RawQuery = values.Encode()

	return url.String(), nil
}

func (c *Client) Stop() {
	c.gatewayWebsocket.Close(websocket.StatusInternalError, "")
	c.cancel()

	c.hbService.Stop()
}
