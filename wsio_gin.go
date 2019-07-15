package wsio

import (
	"sync"

	"github.com/gin-gonic/gin"
	"golang.org/x/net/websocket"
)

type server struct {
	events  Events
	gin     *gin.Engine
	gconn   *conn
	clients map[string]*websocket.Conn
	wx      *sync.RWMutex
}
type conn struct {
	flidx string
}

func serve(events Events, addr, path string, sslCert ...string) error {

	s := &server{}
	s.events = events
	s.clients = make(map[string]*websocket.Conn)
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

		for {
			var in = make([]byte, 0xFFFF)

			n, err := conn.Read(in)
			in = append([]byte{}, in[:n]...)
			if err != nil {
				return
			}

			if s.events.Unpack != nil {
				ctx, flag, _ := s.events.Unpack(in)
				if flag != "" {
					s.wx.Lock()
					s.clients[flag] = conn
					s.wx.Unlock()
				}
				if ctx != nil {
					s.events.Ctx <- &ctx
				}
			}

		}

	})
	h.ServeHTTP(c.Writer, c.Request)
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
