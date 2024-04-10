package annotation

import (
	"github.com/expgo/ag/api"
	"github.com/expgo/factory"
	"go/ast"
	"strings"
)

// @singleton
type Factory struct{}

func (f *Factory) Annotations() map[string][]api.AnnotationType {
	return map[string][]api.AnnotationType{
		AnnotationLog.Val(): {api.AnnotationTypeType},
	}
}

func (f *Factory) New(typedAnnotations []*api.TypedAnnotation) (api.Generator, error) {
	logs := []*Log{}

	for _, ta := range typedAnnotations {
		if ta.Type == api.AnnotationTypeType {
			ts := ta.Node.(*ast.TypeSpec)

			for _, an := range ta.Annotations.Annotations {
				if strings.EqualFold(an.Name, AnnotationLog.Val()) {
					l := factory.New[Log]()
					err := an.To(l)
					if err != nil {
						return nil, err
					}

					l.typeName = ts.Name.Name
					logs = append(logs, l)
				}
			}
		}
	}

	if len(logs) == 0 {
		return nil, nil
	}

	return newGenerator(logs)
}
