package main

import (
	"sync"
	"github.com/bwmarrin/snowflake"
)

var gwSessMgr = CreateSessionManager()

type SessionManager struct {
	sess_mtx sync.RWMutex

	chansubs map[snowflake.ID]map[*gwSession]struct{}
}

func CreateSessionManager() *SessionManager {
	r := &SessionManager{}
	r.chansubs = new(map[snowflake.ID]map[*gwSession]struct{})
	return r
}

func (sm *SessionManager) ChanSub(s *gwSession, c snowflake.ID) {
	sm.sess_mtx.Lock()
	defer sm.sess_mtx.Unlock()
	n := chansubs[c]
	if n == nil { n = new(map[snowflake.ID]struct{}) }
	n[s] = struct{}{}
	chansubs[c] = n
}

func (sm *SessionManager) ChanUnsub(s *gwSession, c snowflake.ID) {
	sm.sess_mtx.Lock()
	defer sm.sess_mtx.Unlock()
	n := chansubs[c]
	if n == nil { n = new(map[snowflake.ID]struct{}) }
	delete(n, s)
	chansubs[c] = n
}

func (sm *SessionManager) Main() {
	db2 := DB.Msg.Session.New()
	
	_ = db2
}
