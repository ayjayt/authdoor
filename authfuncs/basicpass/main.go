package basicpass

import (
	"net/http"

	"github.com/ayjayt/authdoor"
)

type BasicPass struct {
	password string
}

func (b *BasicPass) Check(w http.ResponseWriter, r *http.Request) (authdoor.AuthFuncReturn, error) {
	ret = &authdoor.AuthFuncReturn{
		Auth: AuthFailed,
		Resp: Answered,
		Info: authdoor.InstanceReturnInfo{},
	}
	w.Write("this")
	return ret
}
