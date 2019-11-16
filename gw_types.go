package main

import (
	"github.com/bwmarrin/snowflake"
)

// REST types

type responseGetGateway struct {
	URL string `json:"url"`
}

// Gateway opcodes
const (
	GW_OP_HELLO = 10
	GW_OP_IDENTIFY = 2
	GW_OP_DISPATCH = 0
)

// Gateway events
const (
	GW_EVT_READY = "ready"
)

// Packet received or sent over the gateway websocket.
type gwPacket struct {
	Op int `json:"op"`
	Data interface{} `json:"d"`
	Seq int `json:"s"`
	Type string `json:"t"`
}

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

// OP_HELLO packet data
type gwPktDataHello struct {
	// Time in milliseconds
	HeartbeatInterval int `json:"heartbeat_interval"`
}

// OP_UPDATE_STATUS packet data
type gwPktDataUpdateStatus struct {
	// TODO
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

// Ready event packet data
type gwEvtDataReady struct {
	Version int `json:"v"`
	User User `json:"user"`
	// Discord docs say this is empty. Why is it even here?
	PrivateChannels []interface{} `json:"private_channels"`
	Guilds []UnavailableGuild `json:"guilds"`
	SessionID snowflake.ID
}
