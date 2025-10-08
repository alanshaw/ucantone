package datamodel

type ParseErrorModel struct {
	Name    string `cborgen:"name"`
	Message string `cborgen:"message"`
	Source  string `cborgen:"source"`
	Column  int64  `cborgen:"column"`
	Token   string `cborgen:"token"`
}

func (pe ParseErrorModel) Error() string {
	return pe.Message
}

var _ error = (*ParseErrorModel)(nil)

type ResolutionErrorModel struct {
	Name    string   `cborgen:"name"`
	Message string   `cborgen:"message"`
	At      []string `cborgen:"at"`
}

func (u ResolutionErrorModel) Error() string {
	return u.Message
}

var _ error = (*ResolutionErrorModel)(nil)
