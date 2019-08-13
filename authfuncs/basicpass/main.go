package basicpass

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"time"

	"github.com/ayjayt/authdoor"
	"github.com/cornelk/hashmap"
	"github.com/google/uuid"
)

const (
	form1 = `<html>
	<body><form id="form-`
	form2 = `">
		<input name="password" type=password />
		<input name="reference" type="hidden" value="`
	form3 = `">
		<button type="button" id="submit-`
	form4 = `">Submit</button>
	</form>`
	script1 = `<script type="text/javascript">
	window.addEventListener('DOMContentLoaded', function(e) {	
		myUrl = window.location.href
		document.getElementById("submit-`
	script2 = `").addEventListener('click', function(e) {
			req = new XMLHttpRequest()
			req.timeout = 1200
			req.onerror = function() {
				alert("an error occured")
			}
			req.addEventListener("load", function() {
				location.reload()		
			})
			req.open("POST", myUrl)
			req.send(new FormData(document.getElementById("form-`
	script3 = `")))
			e.StopPropogation()
			return false;
		})
	})
</script></body></html>`
)

// BasicPass supplies an authfunc receiver and stores information to be used by that receiver
type BasicPass struct {
	// Password is the correct password
	Password string
	uuid     string
	form     []byte
	sessions *hashmap.HashMap
}

// New returns a new BasicPass
func New(password string) BasicPass {
	ret := BasicPass{
		Password: password,
		uuid:     uuid.New().String(),
		sessions: &hashmap.HashMap{},
	}
	ret.form = []byte(form1 + ret.uuid + form2 + ret.uuid + form3 + ret.uuid + form4 + script1 + ret.uuid + script2 + ret.uuid + script3)
	return ret
}

// Check is an authfunc that determines whether or a user is authenticated or helps them authenticate
func (b *BasicPass) Check(w http.ResponseWriter, r *http.Request) (authdoor.AuthFuncReturn, error) {
	dump, _ := httputil.DumpRequest(r, true)
	fmt.Printf("Body:%v\n", string(dump))
	fmt.Printf("Method:%v\n", r.Method)
	failure := authdoor.AuthFuncReturn{
		Auth: authdoor.AuthFailed,
		Resp: authdoor.Answered,
		Info: authdoor.InstanceReturnInfo{},
	}
	success := authdoor.AuthFuncReturn{
		Auth: authdoor.AuthGranted,
		Resp: authdoor.Ignored,
		Info: authdoor.InstanceReturnInfo{},
	}
	cookie, err := r.Cookie("basicpass-" + b.uuid)
	if err == nil { // Cookies exists
		fmt.Printf("Cookie exists\n")
		sessionTimeIface, ok := b.sessions.Get(cookie.Value)
		if ok { // Found session
			sessionTime := sessionTimeIface.(time.Time)
			fmt.Printf("Session exists\n")
			if time.Now().Before(sessionTime) {
				fmt.Printf("Authorized\n")
				b.sessions.Set(cookie, time.Now().Add(time.Hour*6))
				return success, nil
			}
		}
	}
	if r.Method == "POST" {
		r.ParseMultipartForm(256)
		if r.Form["reference"][0] == b.uuid && r.Form["password"][0] == b.Password {
			fmt.Printf("Password Accepted\n")
			sess := uuid.New().String()
			http.SetCookie(w, &http.Cookie{
				Name:  "basicpass-" + b.uuid,
				Value: sess,
			})

			b.sessions.Set(sess, time.Now().Add(time.Hour*6))
			success.Resp = authdoor.Answered
			return success, nil
		} else {
			fmt.Printf("Password not accepted\n")
			w.Write([]byte("no\n"))
			return failure, nil
		}
	}
	fmt.Printf("Wrote login\n")
	w.Write(b.form)
	w.Write([]byte("\n"))
	return failure, nil
}
