package wsio

import (
	"github.com/gin-gonic/gin"
	"golang.org/x/net/websocket"
)

type server struct {
	events  Events
	gin     *gin.Engine
	gconn   *conn
	clients map[string]*websocket.Conn
}
type conn struct {
	flidx string
}

func serve(events Events, addr, path string) error {

	s := &server{}
	s.events = events
	s.clients = make(map[string]*websocket.Conn)
	s.gin = gin.Default()
	s.gin.GET(path, s.ginHandler)
	if events.NumLoops == 0 {
		go loopSendConn(s)
	}
	for i := 0; i < events.NumLoops; i++ {
		go loopSendConn(s)
	}
	return s.gin.Run(addr)

}
func (s *server) ginHandler(c *gin.Context) {
	h := websocket.Handler(func(conn *websocket.Conn) {

		for {
			var in = make([]byte, 512)
			_, err := conn.Read(in)
			if err != nil {
				return
			}

			if s.events.Unpack != nil {
				ctx, flag, _ := s.events.Unpack(in)
				if flag != "" {
					s.clients[flag] = conn
				}
				s.events.Ctx <- &ctx

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

			if c, ok := s.clients[*flag]; ok {

				c.Write(*msg)

			}
		}

	}

}
