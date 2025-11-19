package types

import "github.com/alanshaw/ucantone/ucan/promise"

type MessageSendArguments struct {
	To      []string
	Subject string
	Message string
}

type PromisedMsgSendArguments struct {
	From    string
	To      promise.AwaitOK
	Message string
}

type EmailsListArguments struct {
	Limit uint64
}
