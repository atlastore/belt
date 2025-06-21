package logx

// Transport abstracts sending a log entry to an external system.
type Transport interface {
	Send(entry LogEntry) error
}