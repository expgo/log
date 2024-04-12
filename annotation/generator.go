package annotation

import (
	"bytes"
	"fmt"
	"github.com/expgo/ag/api"
	"github.com/expgo/generic/stream"
	"io"
	"strings"
)

type Generator struct {
	logs []*Log
}

func (g *Generator) GetImports() []string {
	return []string{"github.com/expgo/log"}
}

func (g *Generator) WriteConst(wr io.Writer) error {
	buf := bytes.NewBuffer([]byte{})

	buf.WriteString("var (\n")

	for _, l := range g.logs {
		buf.WriteString(fmt.Sprintf(`%s = log.FactoryWithTypePathAndConfigPath("%s", "%s")`, l.Name, l.typePath, l.CfgPath) + "\n")
	}

	buf.WriteString(")\n")

	_, err := io.Copy(wr, buf)
	return err
}

func (g *Generator) WriteInitFunc(wr io.Writer) error {
	return nil
}

func (g *Generator) WriteBody(wr io.Writer) error {
	return nil
}

func newGenerator(logs []*Log) (api.Generator, error) {
	sorted := stream.Must(stream.Of(logs).Sort(func(x, y *Log) int { return strings.Compare(x.Name, y.Name) }).ToSlice())
	return &Generator{sorted}, nil
}
