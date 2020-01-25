package main

import (
	"time"

	"github.com/bwmarrin/snowflake"
	"github.com/mitchellh/mapstructure"
)

// REST types

type responseGetGateway struct {
	URL string `json:"url"`
}

// Gateway opcodes
const (
	GW_OP_HEARTBEAT_ACK = 11
	GW_OP_HELLO         = 10
	GW_OP_UPDATE_STATUS = 3
	GW_OP_IDENTIFY      = 2
	GW_OP_HEARTBEAT     = 1
	GW_OP_DISPATCH      = 0
)

// Gateway events
const (
	GW_EVT_READY = "READY"

	GW_EVT_CHANNEL_CREATE      = "CHANNEL_CREATE"
	GW_EVT_CHANNEL_UPDATE      = "CHANNEL_UPDATE"
	GW_EVT_CHANNEL_DELETE      = "CHANNEL_DELETE"
	GW_EVT_CHANNEL_PINS_UPDATE = "CHANNEL_PINS_UPDATE"

	GW_EVT_MESSAGE_CREATE = "MESSAGE_CREATE"
	GW_EVT_MESSAGE_UPDATE = "MESSAGE_UPDATE"
	GW_EVT_MESSAGE_DELETE = "MESSAGE_DELETE"
	GW_EVT_MESSAGE_ACK    = "MESSAGE_ACK" // Undocumented

	GW_EVT_MESSAGE_REACTION_ADD = "MESSAGE_REACTION_ADD"

	GW_EVT_GUILD_CREATE = "GUILD_CREATE"
	GW_EVT_GUILD_UPDATE = "GUILD_UPDATE"
	GW_EVT_GUILD_DELETE = "GUILD_DELETE"

	GW_EVT_GUILD_MEMBER_ADD    = "GUILD_MEMBER_ADD"
	GW_EVT_GUILD_MEMBER_UPDATE = "GUILD_MEMBER_UPDATE"
	GW_EVT_GUILD_MEMBER_REMOVE = "GUILD_MEMBER_REMOVE"

	GW_EVT_GUILD_ROLE_CREATE = "GUILD_EVT_ROLE_CREATE"
	GW_EVT_GUILD_ROLE_UPDATE = "GUILD_EVT_ROLE_UPDATE"
	GW_EVT_GUILD_ROTE_DELETE = "GUILD_EVT_ROLE_DELETE"

	GW_EVT_PRESENCE_UPDATE = "PRESENCE_UPDATE"
)

// OP_UPDATE_STATUS types
const (
	STATUS_ONLINE    = "online"
	STATUS_OFFLINE   = "offline"
	STATUS_DND       = "dnd"
	STATUS_IDLE      = "idle"
	STATUS_INVISIBLE = "invisible"
	STATUS_UNKNOWN   = ""
)

// Packet received or sent over the gateway websocket.
type gwPacket struct {
	Op   int         `json:"op"`
	Data interface{} `json:"d"`
	Type string      `json:"t"`
	Seq  int         `json:"s"`

	PvtData interface{} `json:"-"`
}

// D decodes the data into the specified packet
func (p *gwPacket) D(o interface{}) { mapstructure.Decode(p.Data, o) }

// mkGwPkt makes a gateway packet with the specified properties
func mkGwPkt(op int, data interface{}, therest ...interface{}) *gwPacket {
	p := &gwPacket{Op: op, Data: data}
	if len(therest) > 0 {
		p.Seq = therest[0].(int)
	}
	if len(therest) > 1 {
		p.Type = therest[1].(string)
	}
	return p
}

// Simpler packet without seq/type
type gwPktMini struct {
	Op   int         `json:"op"`
	Data interface{} `json:"d"`
}

// D decodes the data into the specified packet
func (p *gwPktMini) D(o interface{}) { mapstructure.Decode(p.Data, o) }

// OP_HELLO packet data
type gwPktDataHello struct {
	// Time in milliseconds
	HeartbeatInterval int `json:"heartbeat_interval"`
}

// OS properties in OP_IDENTIFY packet
type _gw_OSProperties struct {
	OS               string `json:"$os"`
	Browser          string `json:"$browser"`
	Device           string `json:"$device"`
	JangleCompatible bool   `json:"love_jangle"`
}

// OP_IDENTIFY packet data
type gwPktDataIdentify struct {
	Token              string                 `json:"token"`
	Properties         _gw_OSProperties       `json:"properties"`
	Compress           bool                   `json:"compress"`
	LargeThreshold     int                    `json:"large_threshold"`
	GuildSubscriptions bool                   `json:"guild_subscriptions"`
	Shard              []int                  `json:"shard"`
	Presence           *gwPktDataUpdateStatus `json:"presence"`
}

// OP_UPDATE_STATUS packet data
type gwPktDataUpdateStatus struct {
	Since  time.Duration `json:"since"`
	Game   interface{}   `json:"game"`
	Status string        `json:"status"`
	AFK    bool          `json:"afk"`
}

// Ready event packet data
type gwEvtDataReady struct {
	Version int          `json:"v"`
	User    *APITypeUser `json:"user"`
	// Discord docs say this is empty. Why is it even here?
	// Scratch that. They lie. This is used for non-bot accounts.
	// It's exactly what it says.
	PrivateChannels interface{}         `json:"private_channels"`
	Guilds          []*UnavailableGuild `json:"guilds"`
	SessionID       snowflake.ID        `json:"session_id"`

	Presences     interface{}   `json:"presences,omitempty"`
	Relationships interface{}   `json:"relationships,omitempty"`
	UserSettings  *UserSettings `json:"user_settings,omitempty"`
	Notes         interface{}   `json:"notes,omitempty"`
}

// Message delete event packet data
type gwEvtDataMessageDelete struct {
	ID        snowflake.ID `json:"id"`
	ChannelID snowflake.ID `json:"channel_id"`
	GuildID   snowflake.ID `json:"guild_id,omitempty"`
}

// Presence update event packet data
type gwEvtDataPresenceUpdate struct {
	User         *APITypeUser   `json:"user"`
	Roles        []snowflake.ID `json:"roles"`
	Game         interface{}    `json:"game"`
	GuildID      snowflake.ID   `json:"guild_id"`
	Status       string         `json:"status"`
	Activities   []interface{}  `json:"activities"`
	ClientStatus interface{}    `json:"client_status"`
	Nick         string         `json:"nick,omitempty"`
}
