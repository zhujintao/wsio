package wsio

type Action int

const (
	None Action = iota

	Detach

	Close

	Shutdown
)

type Events struct {
	NumLoops int
	Unpack   func(in []byte) (ctx interface{}, clientFlag string, action Action)
	Sender   sender
	Ctx      chan *interface{}
}
type sender struct {
	ToChan  chan *string
	MsgChan chan *[]byte
}

func Server(events Events, addr, path string, sslCert ...string) error {

	return serve(events, addr, path, sslCert...)
}
func (e *Events) Send(to string, msg []byte) {

	e.Sender.ToChan <- &to
	e.Sender.MsgChan <- &msg
}
