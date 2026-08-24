package datamodel

type ErrorModel struct {
	ErrorName string `cborgen:"name"`
	Message   string `cborgen:"message"`
}

func (em ErrorModel) Name() string {
	return em.ErrorName
}

func (em ErrorModel) Error() string {
	return em.Message
}

var _ error = (*ErrorModel)(nil)
