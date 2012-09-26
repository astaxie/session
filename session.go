package sessionmanager

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"
)

type Session interface {
	Set(key, value interface{}) bool //set session value
	Get(key interface{}) interface{} //get session value
	Del(key interface{}) bool        //delete session value
}

type Provider interface {
	SessionInit(sid string) (Session, error)
	SessionRead(sid string) (Session, error)
	SessionDestroy(sid string) bool
	SessionGC(maxlifetime int64)
}

var provides = make(map[string]Provider)

// Register makes a session provide available by the provided name.
// If Register is called twice with the same name or if driver is nil,
// it panics.
func Register(name string, provide Provider) {
	if provide == nil {
		panic("session: Register provide is nil")
	}
	if _, dup := provides[name]; dup {
		panic("session: Register called twice for provide " + name)
	}
	provides[name] = provide
}

type SessionManager struct {
	cookieName  string     //private cookiename
	lock        sync.Mutex // protects session
	provider    Provider
	maxlifetime int64
}

func NewSessionManager(provideName, cookieName string, maxlifetime int64) (*SessionManager, error) {
	provider, ok := provides[provideName]
	if !ok {
		return nil, fmt.Errorf("session: unknown provide %q (forgotten import?)", provideName)
	}
	return &SessionManager{provider: provider, cookieName: cookieName, maxlifetime: maxlifetime}, nil
}

//get Session
func (this *SessionManager) SessionStart(w http.ResponseWriter, r *http.Request) (session Session) {
	this.lock.Lock()
	defer this.lock.Unlock()
	cookie, err := r.Cookie(this.cookieName)
	if err != nil || cookie.Value == "" {
		sid := this.sessionId()
		session, _ = this.provider.SessionInit(sid)
		cookie := http.Cookie{Name: this.cookieName, Value: url.QueryEscape(sid), Path: "/"}
		http.SetCookie(w, &cookie)
	} else {
		sid, _ := url.QueryUnescape(cookie.Value)
		session, _ = this.provider.SessionRead(sid)
	}
	return
}

//Destroy sessionid
func (this *SessionManager) SessionDestroy(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(this.cookieName)
	if err != nil || cookie.Value == "" {
		return
	} else {
		this.lock.Lock()
		defer this.lock.Unlock()
		this.provider.SessionDestroy(cookie.Value)
		expiration := time.Now()
		cookie := http.Cookie{Name: this.cookieName, Expires: expiration}
		http.SetCookie(w, &cookie)
	}
}

func (this *SessionManager) GC() {
	this.lock.Lock()
	defer this.lock.Unlock()
	this.provider.SessionGC(this.maxlifetime)
	time.AfterFunc(time.Duration(this.maxlifetime), func() { this.GC() })
}

func (this *SessionManager) sessionId() string {
	b := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return ""
	}
	return base64.URLEncoding.EncodeToString(b)
}
