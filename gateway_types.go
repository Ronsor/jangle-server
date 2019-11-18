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
	GW_OP_HELLO = 10
	GW_OP_IDENTIFY = 2
	GW_OP_HEARTBEAT = 1
	GW_OP_DISPATCH = 0
)

// Gateway events
const (
	GW_EVT_READY = "READY"
)

// OP_UPDATE_STATUS types
type StatusType string
const (
	STATUS_ONLINE = StatusType("online")
	STATUS_OFFLINE = StatusType("offline")
	STATUS_DND = StatusType("dnd")
	STATUS_IDLE = StatusType("idle")
	STATUS_INVISIBLE = StatusType("invisible")
)

// Packet received or sent over the gateway websocket.
type gwPacket struct {
	Op int `json:"op"`
	Data interface{} `json:"d"`
	Type string `json:"t"`
	Seq int `json:"s"`
}

// D decodes the data into the specified packet
func (p *gwPacket) D(o interface{}) { mapstructure.Decode(p.Data, o) }

// mkGwPkt makes a gateway packet with the specified properties
func mkGwPkt(op int, data interface{}, therest ...interface{}) *gwPacket {
	p := &gwPacket{Op:op,Data:data}
	if len(therest) > 0 {
		p.Seq = therest[0].(int)
	}
	if len(therest) > 1 {
		p.Type = therest[0].(string)
	}
	return p
}

// Simpler packet without seq/type
type gwPktMini struct {
	Op int `json:"op"`
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
	OS string `json:"$os"`
	Browser string `json:"$browser"`
	Device string `json:"$device"`
	JangleCompatible bool `json:"love_jangle"`
}

// OP_IDENTIFY packet data
type gwPktDataIdentify struct {
	Token string `json:"token"`
	Properties _gw_OSProperties `json:"properties"`
	Compress bool `json:"compress"`
	LargeThreshold int `json:"large_threshold"`
	GuildSubscriptions bool `json:"guild_subscriptions"`
	Shard []int `json:"shard"`
	Presence gwPktDataUpdateStatus `json:"presence"`
}

// OP_UPDATE_STATUS packet data
type gwPktDataUpdateStatus struct {
	Since time.Duration `json:"since"`
	Game interface{} `json:"game"`
	Status StatusType `json:"status"`
	AFK bool `json:"afk"`
}

// Ready event packet data
type gwEvtDataReady struct {
	Version int `json:"v"`
	User *User `json:"user"`
	// Discord docs say this is empty. Why is it even here?
	PrivateChannels []interface{} `json:"private_channels"`
	Guilds []*UnavailableGuild `json:"guilds"`
	SessionID snowflake.ID `json:"session_id"`
}

