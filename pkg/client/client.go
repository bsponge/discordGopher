package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"runtime"
	"strings"
	"sync"

	"github.com/bsponge/discordGopher/pkg/config"
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

	tokenURL       = "https://discord.com/api/oauth2/token"
	oauth2TokenURL = "https://discord.com/api/oauth2/token"
	apiEndpoint    = "https://discord.com/api/v10"

	playCommand = "play"
)

var mentionRegex = regexp.MustCompile("<@.*>")

type Client struct {
	parentCtx context.Context
	ctx       context.Context
	cancel    context.CancelFunc

	mtx sync.Mutex

	cfg *config.Config

	sequence int

	gatewayWebsocket *websocket.Conn
	resumeGatewayURL *url.URL
	sessionID        string
	guildID          string
	userID           string

	hbService *heartbeatService

	voiceStates                  map[string]object.VoiceState
	voiceServerInformationWaitCh chan struct{}

	voiceClient *voiceClient

	guild *object.Guild
}

func NewClient() (*Client, error) {
	cfg, err := config.LoadConfig("")
	if err != nil {
		return nil, err
	}

	client := &Client{
		cfg: cfg,
	}

	hbService := NewHeartbeatService(client)
	client.hbService = hbService

	return client, nil
}

func (c *Client) Start(ctx context.Context) error {
	c.parentCtx = ctx

	log.Logger().Info("Starting the client")

	if err := c.start(ctx, false); err != nil {
		return err
	}

	log.Logger().Info("The clietn has started successfully")

	return nil
}

func (c *Client) start(ctx context.Context, reconnecting bool) error {
	c.ctx, c.cancel = context.WithCancel(ctx)

	c.voiceStates = make(map[string]object.VoiceState)

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

	if reconnecting {
		event := object.Event[object.Resume]{
			Op: 6,
			D: object.Resume{
				Token:     c.cfg.Token,
				SessionID: c.sessionID,
				Sequence:  c.GetSequence(),
			},
		}

		err := wsjson.Write(c.ctx, ws, event)
		if err != nil {
			return err
		}
	}

	go c.poolMessages(reconnecting)

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

func (c *Client) poolMessages(resuming bool) {
	for {
		_, body, err := c.gatewayWebsocket.Read(c.ctx)
		var closeError websocket.CloseError
		switch {
		case errors.As(err, &closeError):
			log.Logger().WithError(err).Error("The websocket connection was closed")
			shouldReconnect, ok := object.ReconnectOnError[int(closeError.Code)]
			if !ok || !shouldReconnect {
				return
			}

			c.resumeConnection()
			return
		case err != nil:
			log.Logger().WithError(err).Error("Could not read message from gateway wss")
			return
		default:
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
		case 7: // Reconnect
			c.resumeConnection()
			return
		case 9: // Invalid session
			if resp.GetBool("d") {
				c.resumeConnection()
				return
			}
		case 10: // Hello
			heartbeatInterval := resp.Get("d").GetInt("heartbeat_interval")
			err := c.hbService.Start(c.ctx, c.gatewayWebsocket, heartbeatInterval, resuming)
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
		return c.handleReady(payload)
	case object.GuildCreateType:
		return c.handleGuildCreate(payload)
	case object.MessageCreateType:
		return c.handleMessageCreate(payload)
	case object.VoiceStateUpdateType:
		return c.handleVoiceStateUpdate(payload)
	case object.VoiceServerUpdateType:
		return c.handleVoiceServerUpdate(payload)
	default:
		log.Logger().WithField("dispatch_type", dispatch).Warn("Received unknown dispatch")
	}

	return nil
}

func (c *Client) handleReady(payload []byte) error {
	var ready object.Ready
	err := json.Unmarshal(payload, &ready)
	if err != nil {
		return err
	}

	resumeURL, err := url.Parse(ready.ResumeGatewayURL)
	if err != nil {
		return err
	}

	c.resumeGatewayURL = resumeURL
	c.sessionID = ready.SessionID
	c.guildID = ready.Guilds[0].ID
	c.userID = ready.User.ID

	return nil
}

func (c *Client) handleGuildCreate(payload []byte) error {
	var guild object.Guild
	err := json.Unmarshal(payload, &guild)
	if err != nil {
		return err
	}

	c.setGuild(&guild)
	if guild.VoiceStates != nil {
		for _, voiceState := range *guild.VoiceStates {
			c.voiceStates[voiceState.UserID] = voiceState
		}
	}

	return nil
}

func (c *Client) handleMessageCreate(payload []byte) error {
	var message object.Message
	err := json.Unmarshal(payload, &message)
	if err != nil {
		return err
	}

	if message.Content != nil {
		content := mentionRegex.ReplaceAllString(*message.Content, "")
		words := strings.Split(content, " ")
		var filteredWords []string

		for _, word := range words {
			word = strings.TrimSpace(word)
			if word == "" {
				continue
			}

			filteredWords = append(filteredWords, strings.ToLower(word))
		}

		if len(filteredWords) == 0 {
			return nil
		}

		command := filteredWords[0]

		switch command {
		case playCommand:
			c.voiceClient = NewVoiceClient(c.ctx, c)

			voiceState, ok := c.voiceStates[message.Author.ID]
			if !ok {
				return fmt.Errorf("could not find voice state information for user %s", message.Author.Username)
			}

			log.Logger().Info("VOICESTATE ", voiceState)

			c.voiceClient.ConnectToVoiceChannel(c.guildID, *voiceState.ChannelID, false, false)
		default:
			log.Logger().WithField("command", command).WithField("user", message.Author.Username).Info("User used unknown command")
		}
	}

	return nil
}

func (c *Client) handleVoiceStateUpdate(payload []byte) error {
	var voiceState object.VoiceState
	err := json.Unmarshal(payload, &voiceState)
	if err != nil {
		return err
	}

	c.voiceStates[voiceState.UserID] = voiceState

	if voiceState.UserID == c.userID {
		c.voiceServerInformationWaitCh <- struct{}{}
	}

	return nil
}

func (c *Client) handleVoiceServerUpdate(payload []byte) error {
	var voiceServerUpdate object.VoiceServerUpdate
	err := json.Unmarshal(payload, &voiceServerUpdate)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) resumeConnection() {
	c.Stop()
	c.ctx, c.cancel = context.WithCancel(c.parentCtx)

	log.Logger().Info("Reconnecting...")

	err := c.start(c.ctx, true)
	if err != nil {
		log.Logger().WithError(err).Error("Could not resume connection")
	}
}

func (c *Client) getGuild() *object.Guild {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	return c.guild
}

func (c *Client) setGuild(guild *object.Guild) {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	c.guild = guild
}

func (c *Client) Identify() error {
	var intent int = 1         // GUILDS
	intent = intent | (1 << 1) // GUILD_MEMBERS
	intent = intent | (1 << 7) // GUILD_VOICE_STATES
	intent = intent | (1 << 8) // GUILD_PRESENCES
	intent = intent | (1 << 9) // GUILD_MESSAGES

	identify := object.Identify{
		Token: c.cfg.Token,
		Properties: map[string]any{
			"os":      runtime.GOOS,
			"browser": "discordGopher",
			"device":  "discordGopher",
		},
		Compress: false,
		Intents:  intent,
	}

	event := object.Event[object.Identify]{
		Op: 2,
		D:  identify,
	}

	err := wsjson.Write(c.ctx, c.gatewayWebsocket, event)
	if err != nil {
		return err
	}

	return nil
}

// getGateway gets gateway WSS URL which is used to listen for discord server events.
func (c *Client) getGatewayURL() (string, error) {
	if c.resumeGatewayURL != nil {
		return c.resumeGatewayURL.String(), nil
	}

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/gateway/bot", apiEndpoint), nil)
	if err != nil {
		return "", err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bot %s", c.cfg.Token))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

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
	log.Logger().Info("Stopping the client")

	c.gatewayWebsocket.Close(websocket.StatusInternalError, "")
	c.cancel()

	c.hbService.Stop()

	log.Logger().Info("The client has stopped")
}
