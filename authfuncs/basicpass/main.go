package basicpass

import (
	"net/http"

	"github.com/ayjayt/authdoor"
	"github.com/google/uuid"
)

const (
	form1 = `<html><body><form>
	<input id="form-`
	form2 = `" type=password />
	<button id="submit-`
	form3 = `">Submit</button>
</form>`
	script1 = `<script type="text/javascript">
	document.getElementById("submit-`
	script2 = `").onclick = function(e) {
		req = new XMLHttpRequest()
		req.timeout = 1200
		req.onerror = function() {
			alert("an error occured")
		}
		req.addEventListener("load", function() {
			location.reload()		
		})
		req.open("POST", window.location.href)
		// post body
		req.send("basicpass-`
	script3 = `:"+document.getElementById("form-`
	script4 = `"),value)
	}
</script></body></html>`
)

// BasicPass supplies an authfunc receiver and stores information to be used by that receiver
type BasicPass struct {
	Password string
	uuid     string
	form     []byte
}

// New returns a new BasicPass
func New(password string) BasicPass {
	ret := BasicPass{
		Password: password,
		uuid:     uuid.New().String(),
	}
	ret.form = []byte(form1 + ret.uuid + form2 + ret.uuid + form3 + script1 + ret.uuid + script2 + ret.uuid + script3 + ret.uuid + script4)
	return ret
}

// Check is an authfunc that determines whether or a user is authenticated or helps them authenticate
func (b *BasicPass) Check(w http.ResponseWriter, r *http.Request) (authdoor.AuthFuncReturn, error) {
	failure := authdoor.AuthFuncReturn{
		Auth: authdoor.AuthFailed,
		Resp: authdoor.Answered,
		Info: authdoor.InstanceReturnInfo{},
	}
	if r.Method != "POST" {
		w.Write(b.form)
		w.Write([]byte("\n"))
		return failure, nil
	}
	success := authdoor.AuthFuncReturn{
		Auth: authdoor.AuthGranted,
		Resp: authdoor.Ignored,
		Info: authdoor.InstanceReturnInfo{},
	}
	return success, nil
}
