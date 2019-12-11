package agent

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"

	"github.com/gorilla/mux"

	"github.com/ribice/goch"

	"github.com/gorilla/websocket"
	"github.com/ribice/goch/internal/broker"
)

var (
	alfaRgx *regexp.Regexp
)

// NewAPI creates new websocket api
func NewAPI(m *mux.Router, br *broker.Broker, store ChatStore, lim Limiter) *API {
	api := API{
		broker: br,
		store:  store,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin:     func(r *http.Request) bool { return true },
		},
	}
	alfaRgx = regexp.MustCompile("^[a-zA-Z0-9_]*$")

	m.HandleFunc("/connect", api.connect).Methods("GET")

	return &api
}

// API represents websocket api service
type API struct {
	broker   *broker.Broker
	store    ChatStore
	upgrader websocket.Upgrader
	rlim     Limiter
}

// Limiter represents chat service limit checker
type Limiter interface {
	ExceedsAny(map[string]goch.Limit) error
}

func (api *API) connect(w http.ResponseWriter, r *http.Request) {
	fmt.Println("start aaaa0")
	conn, err := api.upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, fmt.Sprintf("error while upgrading to ws connection: %v", err), 500)
		return
	}
	fmt.Println("start aaaa1")
	req, err := api.waitConnInit(conn)
	fmt.Println("start aaaa2")
	if err != nil {
		fmt.Println("start err")
		fmt.Println(err)
		if err == errConnClosed {
			return
		}
		writeErr(conn, err.Error())
		return
	}
	fmt.Println("start aaaa")
	agent := New(api.broker, api.store)
	agent.HandleConn(conn, req)
}

type initConReq struct {
	Channel string  `json:"channel"`
	UID     string  `json:"uid"`
	Secret  string  `json:"secret"` // User secret
	LastSeq *uint64 `json:"last_seq"`
}

func (api *API) bindReq(r *initConReq) error {
	fmt.Println("??start bindReq")
	fmt.Println("??secret")
	fmt.Println(r.Secret)
	fmt.Println("??channel")
	fmt.Println(r.Channel)
	fmt.Println("??uid")
	fmt.Println(r.UID)
	if !alfaRgx.MatchString(r.Secret) {
		return errors.New("secret must contain only alphanumeric and underscores")
	}
	if !alfaRgx.MatchString(r.Channel) {
		return errors.New("channel must contain only alphanumeric and underscores")
	}
	fmt.Println("??end bindreq")
	return nil
	// return api.rlim.ExceedsAny(map[string]goch.Limit{
	// 	r.UID:     goch.UIDLimit,
	// 	r.Secret:  goch.SecretLimit,
	// 	r.Channel: goch.ChanLimit,
	// })
}

var errConnClosed = errors.New("connection closed")

/* func (api *API) waitConnInit(conn *websocket.Conn) (*initConReq, error) {
	fmt.Println("x1")
	t, wsr, err := conn.NextReader()
	if err != nil || t == websocket.CloseMessage {
		fmt.Println("x2")
		return nil, errConnClosed
	}

	var req *initConReq
	fmt.Println("x3")
	fmt.Println(req)
	fmt.Println(wsr)
	err = json.NewDecoder(wsr).Decode(req)
	if err != nil {
		fmt.Println("x4")
		return nil, err
	}
	fmt.Println("x5")
	if err = api.bindReq(req); err != nil {
		fmt.Println("x6")
		return nil, err
	}
	fmt.Println("x7")
	return req, nil
} */
func (api *API) waitConnInit(conn *websocket.Conn) (*initConReq, error) {
	// TODO - Add join timeout
	fmt.Println("x1")
	t, wsr, err := conn.NextReader()
	if err != nil || t == websocket.CloseMessage {
		fmt.Println("x2")
		return nil, errConnClosed
	}

	var req initConReq
	fmt.Println("x3")
	// fmt.Println(req)
	// fmt.Println(wsr)
	err = json.NewDecoder(wsr).Decode(&req)
	// fmt.Println("decode")
	// fmt.Println(err)
	if err != nil {
		fmt.Println("x4")
		return nil, err
	}
	fmt.Println("x5")
	if err = api.bindReq(&req); err != nil {
		fmt.Println("x6")
		return nil, err
	}
	fmt.Println("x7")
	return &req, nil
}
