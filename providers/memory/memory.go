package memory

import (
	"container/list"
	"fmt"
	"github.com/astaxie/sessionmanager"
	"sync"
	"time"
)

var d = &Provider{list: list.New()}

type SessionStore struct {
	sid          string                      //session id唯一标示	  	
	timeAccessed time.Time                   //最后访问时间	  	
	value        map[interface{}]interface{} //session里面存储的值
}

func (this *SessionStore) Set(key, value interface{}) bool {
	this.value[key] = value
	d.SessionUpdate(this.sid)
	return true
}

func (this *SessionStore) Get(key interface{}) interface{} {
	d.SessionUpdate(this.sid)
	if v, ok := this.value[key]; ok {
		return v
	} else {
		return nil
	}
	return nil
}

func (this *SessionStore) Del(key interface{}) bool {
	delete(this.value, key)
	d.SessionUpdate(this.sid)
	return true
}

type Provider struct {
	lock     sync.Mutex               //用来锁
	sessions map[string]*list.Element //用来存储在内存
	list     *list.List               //用来做gc
}

func (this *Provider) SessionInit(sid string) (sessionmanager.Session, error) {
	this.lock.Lock()
	defer this.lock.Unlock()
	v := make(map[interface{}]interface{}, 0)
	newsess := &SessionStore{sid: sid, timeAccessed: time.Now(), value: v}
	fmt.Println(newsess)
	element := this.list.PushBack(newsess)
	fmt.Println(element)
	this.sessions[sid] = element
	return newsess, nil
}

func (this *Provider) SessionRead(sid string) (sessionmanager.Session, error) {
	if element, ok := this.sessions[sid]; ok {
		return element.Value.(*SessionStore), nil
	} else {
		sess, err := this.SessionInit(sid)
		return sess, err
	}
	return nil, nil
}

func (this *Provider) SessionDestroy(sid string) bool {
	if element, ok := this.sessions[sid]; ok {
		delete(this.sessions, sid)
		this.list.Remove(element)
		return true
	} else {
		return false
	}
	return true
}

func (this *Provider) SessionGC(maxlifetime int64) {
	this.lock.Lock()
	defer this.lock.Unlock()

	for {
		element := this.list.Back()
		if element == nil {
			break
		}
		if (element.Value.(*SessionStore).timeAccessed.Unix() + maxlifetime) < time.Now().Unix() {
			this.list.Remove(element)
			delete(this.sessions, element.Value.(*SessionStore).sid)
		} else {
			break
		}
	}
}

func (this *Provider) SessionUpdate(sid string) bool {
	this.lock.Lock()
	defer this.lock.Unlock()
	fmt.Println("update")
	fmt.Println(this.sessions)
	if element, ok := this.sessions[sid]; ok {
		fmt.Println("begin moveToFront")
		element.Value.(*SessionStore).timeAccessed = time.Now()
		this.list.MoveToFront(element)
		fmt.Println("end moveToFront")
		return true
	} else {
		return false
	}
	return true
}

func init() {
	d.sessions = make(map[string]*list.Element, 0)
	sessionmanager.Register("memory", d)
}
