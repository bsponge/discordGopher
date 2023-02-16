package client

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/bsponge/discordGopher/pkg/config"
	"github.com/bsponge/discordGopher/pkg/log"

	"github.com/valyala/fastjson"
	"nhooyr.io/websocket"
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

	cfg *config.Config

	gatewayWebsocket *websocket.Conn
}

func NewClient(ctx context.Context) (*Client, error) {
	log.Logger().Trace("Creating the client")

	cfg, err := config.LoadConfig("")
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(ctx)

	return &Client{
		ctx:    ctx,
		cancel: cancel,
		cfg:    cfg,
	}, nil
}

func (c *Client) Start() error {
	log.Logger().Trace("Starting the client")

	gatewayURL, err := c.getGatewayURL()
	if err != nil {
		return err
	}

	ws, _, err := websocket.Dial(c.ctx, gatewayURL, nil)
	if err != nil {
		return err
	}

	c.gatewayWebsocket = ws

	go c.pollMessages()

	log.Logger().Infof("Gateway URL %s", gatewayURL)

	return nil
}

func (c *Client) pollMessages() {
	for {
		_, body, err := c.gatewayWebsocket.Read(c.ctx)
		if err == context.Canceled {
			return
		}
		if err != nil {
			log.Logger().WithError(err).Error("Could not read message from gateway wss")
			continue
		}

		log.Logger().Info(string(body))
	}
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

func (c *Client) Close() {
	c.gatewayWebsocket.Close(websocket.StatusInternalError, "")
	c.cancel()
}
