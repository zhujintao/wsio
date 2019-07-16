package wsio

import (
	"fmt"
	"sync"

	"github.com/gin-gonic/gin"
	"golang.org/x/net/websocket"
)

type server struct {
	events  Events
	gin     *gin.Engine
	gconn   *conn
	clients map[string]*websocket.Conn
	flidxs  map[*websocket.Conn]string
	wx      *sync.RWMutex
}
type conn struct {
	flidx string
}

func serve(events Events, addr, path string, sslCert ...string) error {

	s := &server{}
	s.events = events
	s.clients = make(map[string]*websocket.Conn)
	s.flidxs = make(map[*websocket.Conn]string)
	s.wx = &sync.RWMutex{}
	s.gin = gin.Default()

	s.gin.GET(path, s.ginHandler)
	if events.NumLoops == 0 {
		go loopSendConn(s)
	}
	for i := 0; i < events.NumLoops; i++ {
		go loopSendConn(s)
	}
	if len(sslCert) == 2 {
		return s.gin.RunTLS(addr, sslCert[0], sslCert[1])
	} else {
		return s.gin.Run(addr)
	}

}

func (s *server) ginHandler(c *gin.Context) {

	h := websocket.Handler(func(conn *websocket.Conn) {
		defer s.loopCloseConn(conn)
		for {

			var in = make([]byte, 0xFFFF)

			n, err := conn.Read(in)
			in = append([]byte{}, in[:n]...)
			if err != nil {

				return
			}

			if s.events.Unpack != nil {
				ctx, flag, action := s.events.Unpack(in)
				if action == Close {
					conn.Close()
				}
				if flag != "" {
					s.wx.Lock()
					s.clients[flag] = conn
					s.flidxs[conn] = flag
					s.wx.Unlock()
				}
				if ctx != nil {
					s.events.Ctx <- &ctx
				}
				fmt.Println("[wsio] s.clients MAP length Total: ", len(s.clients))
			}

		}

	})

	h.ServeHTTP(c.Writer, c.Request)
}
func (s *server) loopCloseConn(conn *websocket.Conn) {
	s.wx.RLock()
	flidx := s.flidxs[conn]
	s.wx.RUnlock()
	s.wx.Lock()
	delete(s.clients, flidx)
	s.wx.Unlock()
	conn.Close()
}
func loopSendConn(s *server) {

	for {

		flag := <-s.events.Sender.ToChan
		msg := <-s.events.Sender.MsgChan

		if *flag == "toall" {
			for _, c := range s.clients {
				c.Write(*msg)
			}

		} else {

			s.wx.RLock()
			c, ok := s.clients[*flag]
			s.wx.RUnlock()

			if ok {

				c.Write(*msg)

			}
		}

	}

}
