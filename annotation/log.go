package annotation

/*
	@Enum {
		Log
	}
*/
type Annotation string

type Log struct {
	CfgPath  string `value:"logging"`
	typeName string
}
