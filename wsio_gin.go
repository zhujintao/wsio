package wsio

import (
	"io"

	"github.com/gin-gonic/gin"
	"golang.org/x/net/websocket"
)

type server struct {
	events  Events
	gin     *gin.Engine
	clients map[string]*websocket.Conn
}

func serve(events Events, addr, path string) error {

	s := &server{}
	s.clients = make(map[string]*websocket.Conn)
	s.gin = gin.Default()
	s.gin.GET(path, s.ginHandler)
	return s.gin.Run(addr)

}
func (s *server) ginHandler(c *gin.Context) {
	h := websocket.Handler(func(conn *websocket.Conn) {
		io.Copy(conn, conn)

	})
	h.ServeHTTP(c.Writer, c.Request)

}
