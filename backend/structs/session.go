package structs

import "sync"

type LoginParams struct {
	User string
	Pass string
	ID   string
}

type Session struct {
	Active bool
	User   string
	ID     string
	Group  string
}

var (
	sessionInstance *Session
	sessionOnce     sync.Once
)

func GetSession() *Session {
	sessionOnce.Do(func() {
		sessionInstance = &Session{
			Active: false,
		}
	})
	return sessionInstance
}

func (s *Session) Login(user, id, group string) {
	s.Active = true
	s.User = user
	s.ID = id
	s.Group = group
}

func (s *Session) Logout() {
	s.Active = false
	s.User = ""
	s.ID = ""
	s.Group = ""
}
