package registry

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/ICKelin/gtun/common"
	"github.com/ICKelin/gtun/logs"
)

type GtunConfig struct {
	Listener string   `toml:"listen"`
	Tokens   []string `toml:"tokens"` // 用户授权码
}

type gtun struct {
	listener string
	tokens   []string
	m        *Models
}

func NewGtun(cfg *GtunConfig, m *Models) *gtun {
	return &gtun{
		listener: cfg.Listener,
		tokens:   cfg.Tokens,
		m:        m,
	}
}

func (g *gtun) onGtunAccess(w http.ResponseWriter, r *http.Request) {
	content, err := ioutil.ReadAll(r.Body)
	if err != nil {
		logs.Error("read body fail: %v", err)
		return
	}
	defer r.Body.Close()

	regInfo := &common.C2GRegister{}
	err = json.Unmarshal(content, &regInfo)
	if err != nil {
		bytes := common.Response(nil, err)
		w.Write(bytes)
		return
	}

	if g.checkAuth(regInfo) == false {
		bytes := common.Response(nil, errors.New("auth fail"))
		w.Write(bytes)
		return
	}

	gtund, err := g.m.RandomGetGtund(regInfo.IsWindows)
	if err != nil {
		bytes := common.Response(nil, err)
		w.Write(bytes)
		return
	}

	respObj := &common.G2CRegister{
		ServerAddress: fmt.Sprintf("%s:%d", gtund.PublicIP, gtund.Port),
	}

	bytes := common.Response(respObj, nil)
	w.Write(bytes)
	logs.Info("register from %s", r.RemoteAddr)
}

func (g *gtun) checkAuth(regInfo *common.C2GRegister) bool {
	// just write for...
	for _, token := range g.tokens {
		if token == regInfo.AuthToken {
			return true
		}
	}
	return false
}
