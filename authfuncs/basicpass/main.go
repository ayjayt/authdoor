package basicpass

import (
	"net/http"

	"github.com/ayjayt/authdoor"
)

type BasicPass struct {
	password string
}

func (b *BasicPass) Check(w http.ResponseWriter, r *http.Request) (authdoor.AuthFuncReturn, error) {
	ret := &authdoor.AuthFuncReturn{
		Auth: authdoor.AuthFailed,
		Resp: authdoor.Answered,
		Info: authdoor.InstanceReturnInfo{},
	}
	w.Write([]byte("this"))
	return ret, nil
}
