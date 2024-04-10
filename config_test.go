package log

import (
	"github.com/expgo/factory"
	"github.com/expgo/structure"
	"gopkg.in/natefinch/lumberjack.v2"
	"testing"
)

func TestStructChange(t *testing.T) {
	c := factory.New[Config]()
	log, err := structure.ConvertTo[*lumberjack.Logger](&c.File)
	if err != nil {
		t.Fatalf("Failed to convert LogFile to lumberjack.Logger: %v", err)
	}
	log.Close()
}
