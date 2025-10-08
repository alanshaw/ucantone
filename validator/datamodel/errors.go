package datamodel

type ErrorModel struct {
	Name    string `cborgen:"name"`
	Message string `cborgen:"message"`
}

func (u ErrorModel) Error() string {
	return u.Message
}

var _ error = (*ErrorModel)(nil)
