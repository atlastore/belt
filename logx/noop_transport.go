package logx

var (
	_ Transport = &NOOPTransport{}
)

type NOOPTransport struct {}

func (t *NOOPTransport) Send(_ LogEntry) error {
	return nil
}