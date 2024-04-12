package annotation

/*
	@Enum {
		Log
	}
*/
type Annotation string

type Log struct {
	Name     string `value:"logger"`
	CfgPath  string `value:"logging"`
	typePath string
}
