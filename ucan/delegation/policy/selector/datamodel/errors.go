package datamodel

type ParseErrorModel struct {
	Name    string `cborgen:"name" dagjsongen:"name"`
	Message string `cborgen:"message" dagjsongen:"message"`
	Source  string `cborgen:"source" dagjsongen:"source"`
	Column  int64  `cborgen:"column" dagjsongen:"column"`
	Token   string `cborgen:"token" dagjsongen:"token"`
}

func (pe ParseErrorModel) Error() string {
	return pe.Message
}

var _ error = (*ParseErrorModel)(nil)

type ResolutionErrorModel struct {
	Name    string   `cborgen:"name" dagjsongen:"name"`
	Message string   `cborgen:"message" dagjsongen:"message"`
	At      []string `cborgen:"at" dagjsongen:"at"`
}

func (u ResolutionErrorModel) Error() string {
	return u.Message
}

var _ error = (*ResolutionErrorModel)(nil)
