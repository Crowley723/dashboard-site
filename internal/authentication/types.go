package authentication

type SessionKey string

type OIDCError struct {
	RedirectURL string
	Message     string
}

func (e *OIDCError) Error() string {
	return e.Message
}
