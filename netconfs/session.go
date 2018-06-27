/*****************************************************************************/
/**
* \file       session.go
* \author     gaoxj
* \date       2018/06/05
* \brief      ssh会话管理
* \note       Copyright (c) 2000-2020  赛特斯信息科技股份有限公司
* \remarks    修改日志
******************************************************************************/

package main

import (
	"fmt"
	"log"
	"sync"
)

const (
	ctrlChLen = 1024
)

type SessionMgr struct {
	sync.RWMutex
	sessions map[Sid]*Session
	freeSid  uint64
}

type Session struct {
	id       Sid
	radd     string
	conn     int32
	ctrlDown chan *Message //从rest服务器过来的消息通道
	ctrlUp   chan *Message //从flex GW过来的消息通道
}

type Sid struct {
	sid uint64
}

type Message struct {
	data string
}

func (sm *SessionMgr) Init() {
	sm.sessions = make(map[Sid]*Session)
	sm.freeSid = 1
}

func (sm *SessionMgr) AddSession(raddr string, conn int32) *Session {
	id := sm.GetSid()

	sm.Lock()
	defer sm.Unlock()

	s := &Session{}
	s.radd = raddr
	s.conn = conn
	s.id = id
	s.ctrlUp = make(chan *Message, ctrlChLen)
	s.ctrlDown = make(chan *Message, ctrlChLen)
	sm.sessions[id] = s

	go s.run()
	log.Printf("New session arrived %v", s)
	return s
}

func (sm *SessionMgr) DelSession(s *Session) {
	sm.Lock()
	defer sm.Unlock()

	log.Printf("Del session %v", s)
	if _, ok := sm.sessions[s.id]; ok {
		delete(sm.sessions, s.id)
	}
}

func (sm *SessionMgr) GetSid() Sid {
	sm.Lock()
	defer sm.Unlock()

	sid := sm.freeSid
	sm.freeSid++

	return Sid{sid}
}

func (sm *SessionMgr) FindSession(id Sid) *Session {
	sm.RLock()
	defer sm.RUnlock()
	if s, ok := sm.sessions[id]; ok {
		return s
	}

	return nil
}

func (sm SessionMgr) String() string {
	sm.RLock()
	defer sm.RUnlock()
	var out string

	for ii, s := range sm.sessions {
		out += fmt.Sprintf("%v: %v\n", ii, s)
	}

	return out
}

func NewSession(raddr string, conn int32) *Session {
	sm := new(NetConfSData)
	sm.Init()
	return sm.sessMgr.AddSession(raddr, conn)
}

func DelSession(s *Session) {
	sm := new(NetConfSData)
	sm.Init()
	sm.sessMgr.DelSession(s)
}

func (s *Session) MsgUp(data string) {
	s.ctrlUp <- NewMessage(data)
}

func (s *Session) MsgDown(data string) {
	s.ctrlDown <- NewMessage(data)
}

func (s *Session) run() {
	//defer log.DeferHdl()

	for {
		select {
		//case m, ok := <-s.ctrlUp:
		//if !ok {
		//log.Printf("ctrlup close,exit session %v", s)
		//return
		//}
		//restClientReq(s.id.sid, m.data)
		case m, ok := <-s.ctrlDown:
			if !ok {
				log.Printf("ctrldown close,exit session %v", s)
				return
			}
			err := netConfWrite(s.conn, []byte(m.data))
			if err != nil {
				log.Printf("Write to session %v fail: %s", s, err)
			}
			log.Printf("Write to session %v: %s", s, m.data)
		}
	}
}

func NewMessage(data string) *Message {
	return &Message{data}
}

func pipeMsgDown(sid uint64, devmsg string) {
	local := new(NetConfSData)
	local.Init()
	s := local.sessMgr.FindSession(Sid{sid})
	if s == nil {
		log.Printf("pipeMsgDown:could not find session by sid %d", sid)
		return
	}

	s.MsgDown(devmsg)
}

func pipeMsgUp(sid uint64, devmsg string) {
	local := new(NetConfSData)
	local.Init()
	s := local.sessMgr.FindSession(Sid{sid})
	if s == nil {
		log.Printf("pipeMsgDown:could not find session by sid %d", sid)
		return
	}

	s.MsgUp(devmsg)
}

func fmtSessions() string {
	sm := new(NetConfSData)
	sm.Init()
	return fmt.Sprintf("%v", sm.sessMgr)
}
