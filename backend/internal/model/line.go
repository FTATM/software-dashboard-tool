package model

type LineClient interface {
	VerifyIDToken(idToken string) (string, error)
	SendPushMessage(lineUserID string, messageText string) error
}
