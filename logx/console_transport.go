package logx

import (
	"fmt"
)

var (
	_ Transport = &ConsoleTransport{}
)

type ConsoleTransport struct{}

func (t *ConsoleTransport) Send(entry LogEntry) error {
	fields, err := entry.MarshalFields()
	if err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Printf("[REMOTE] %s: %s, service: %s, id: %s (%v)\n", entry.Level.String(), entry.Message, entry.Service, entry.InstanceID, string(fields))
	return nil
}