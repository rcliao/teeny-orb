package container

import (
	"fmt"
	"io"
	"strings"
	"time"
)

// IDGenerator interface for generating unique IDs (for testability)
type IDGenerator interface {
	GenerateID() string
}

// DefaultIDGenerator generates time-based unique IDs
type DefaultIDGenerator struct{}

func (g *DefaultIDGenerator) GenerateID() string {
	return fmt.Sprintf("teeny-orb-%d", time.Now().UnixNano())
}

// StaticIDGenerator generates predictable IDs for testing
type StaticIDGenerator struct {
	prefix  string
	counter int
}

func NewStaticIDGenerator(prefix string) *StaticIDGenerator {
	return &StaticIDGenerator{prefix: prefix, counter: 0}
}

func (g *StaticIDGenerator) GenerateID() string {
	g.counter++
	return fmt.Sprintf("%s-%d", g.prefix, g.counter)
}

// mapToEnvSlice converts environment map to slice format
func mapToEnvSlice(env map[string]string) []string {
	var envSlice []string
	for k, v := range env {
		envSlice = append(envSlice, fmt.Sprintf("%s=%s", k, v))
	}
	return envSlice
}

// separateOutput separates stdout and stderr from Docker's multiplexed stream
// This is a simplified implementation - in production would properly handle Docker's stream format
func separateOutput(reader io.Reader) (io.Reader, io.Reader) {
	return reader, strings.NewReader("")
}
