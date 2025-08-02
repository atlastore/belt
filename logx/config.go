package logx

import "strings"

// State is the mode the logger is being used in.
type State int

const (
	// The logger is being used in Development mode.
	Development State = iota + 1
	// The logger is being used in Production mode.
	Production
)

func (s State) String() string {
	switch s {
	case Development:
		return "development"
	case Production:
		return "production"
	default:
		return "unknown"
	}
}

// Config is the internal config for setup for the logger.
type Config struct {
	State State
	NumWorkers int
	Service string
	InstanceID string
}

//NewConfig creates a new instance of the Config struct with the provided state, service, instanceID and optional numWorkers.
// The state is which mode the logger should be used with.
// Service is what kind of service is being used
// InstanceID 
func NewConfig(state State, service, instanceID string, numWorkers ...int) Config {
	num := 5
	if len(numWorkers) == 1 {
		num = numWorkers[0]
	}
	return Config{
		State: state,
		Service: service,
		InstanceID: instanceID,
		NumWorkers: num,
	}
}

func (c Config) valid() bool {
	return strings.TrimSpace(c.Service) != "" && strings.TrimSpace(c.InstanceID) != "" && (c.State == Development || c.State == Production)
}