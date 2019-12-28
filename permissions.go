package main

// Permission constants
const (
	PERM_CREATE_INVITE        = 0x1
	PERM_KICK_MEMBERS         = 0x2
	PERM_BAN_MEMBERS          = 0x4
	PERM_ADMINISTRATOR        = 0x8
	PERM_MANAGE_CHANNELS      = 0x10
	PERM_MANAGE_GUILD         = 0x20
	PERM_ADD_REACTIONS        = 0x40
	PERM_READ_AUDIT_LOG       = 0x80
	PERM_VIEW_CHANNEL         = 0x400 // Read Messages
	PERM_SEND_MESSAGES        = 0x800
	PERM_SEND_TTS_MESSAGES    = 0x1000
	PERM_MANAGE_MESSAGES      = 0x2000
	PERM_EMBED_LINKS          = 0x4000
	PERM_ATTACH_FILES         = 0x8000
	PERM_READ_MESSAGE_HISTORY = 0x10000
	PERM_MENTION_EVERYONE     = 0x20000
	PERM_USE_EXTERNAL_EMOJIS  = 0x40000
	// TODO: the rest
	PERM_CHANGE_NICKNAME = 0x4000000
	PERM_EVERYTHING      = 0xFFFFFFFF
)

type PermSet int

func (a PermSet) Has(b PermSet) bool {
	if a&PERM_ADMINISTRATOR != 0 {
		return true
	}
	return a&b != 0
}
