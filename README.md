sessionmanager
==============

sessionmanager is a golang session manager. It can use session for many providers.Just like the `database/sql` and `database/sql/driver`.

##How to install

	go get github.com/astaxie/sessionmanager


##install providers
Now I complete a memory provider. The next I will develop other providers.

	go get github.com/astaxie/sessionmanager/providers/memory

##How do we use it?

first you must import it


	import (
		"github.com/astaxie/sessionmanager"
		_ "github.com/astaxie/sessionmanager/providers/memory"
	)

then in you web app init the globalsession manager
	
	var globalSessions *sessionmanager.SessionManager


	func init() {
		globalSessions, _ = sessionmanager.NewSessionManager("memory", "gosessionid", 3600)
		go globalSessions.GC()
	}


at last in the handlerfunc you can use it like this

	func login(w http.ResponseWriter, r *http.Request) {
		sess := globalSessions.SessionStart(w, r)
		username := sess.Get("username")
		fmt.Println(username)
		if r.Method == "GET" {
			t, _ := template.ParseFiles("login.gtpl")
			t.Execute(w, nil)
		} else {
			fmt.Println("username:", r.Form["username"])
			sess.Set("username", r.Form["username"])
			fmt.Println("password:", r.Form["password"])
		}
	}
	


##How to write own provider
When we develop a web app, maybe you want to write a provider because you must meet the requirements.

Write a provider is so easy. You only define two struct type(Session and Provider),which satisfy the interface definition.Maybe The memory provider is a good example for you.

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