package datamodel

type ErrorModel struct {
	ErrorName string `cborgen:"name"`
	Message   string `cborgen:"message"`
}

func (u ErrorModel) Name() string {
	return u.ErrorName
}

func (u ErrorModel) Error() string {
	return u.Message
}

var _ error = (*ErrorModel)(nil)
