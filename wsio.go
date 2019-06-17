package wsio

type Action int

const (
	None Action = iota

	Detach

	Close

	Shutdown
)

type Events struct {
	Sender sender
	Ctx    chan *interface{}
}
type sender struct {
	ToChan  chan *string
	MsgChan chan *[]byte
}

func Server(events Events, addr, path string) error {

	return serve(events, addr, path)
}
