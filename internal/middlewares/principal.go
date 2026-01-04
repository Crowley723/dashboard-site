package middlewares

//go:generate mockgen -source=principal.go -destination=../mocks/principal.go -package=mocks

type Principal interface {
	GetIss() string
	GetSub() string
	GetScopes() []string
	HasScope(string) bool
	MatchesOwner(iss, sub string) bool
}
