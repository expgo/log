package annotation

import (
	"github.com/expgo/ag/api"
	"github.com/expgo/factory"
	"strings"
)

// @singleton
type Factory struct{}

func (f *Factory) Annotations() map[string][]api.AnnotationType {
	return map[string][]api.AnnotationType{
		AnnotationLog.Val(): {api.AnnotationTypeGlobal},
	}
}

func (f *Factory) New(typedAnnotations []*api.TypedAnnotation) (api.Generator, error) {
	logs := []*Log{}

	for _, ta := range typedAnnotations {
		if ta.Type == api.AnnotationTypeGlobal {
			for _, an := range ta.Annotations.Annotations {
				if strings.EqualFold(an.Name, AnnotationLog.Val()) {
					l := factory.New[Log]()
					err := an.To(l)
					if err != nil {
						return nil, err
					}

					l.typePath = ta.FileInfo.FileFullPath
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
