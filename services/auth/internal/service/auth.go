package service

const (
	ValidUsername = "student"
	ValidPassword = "student"
	ValidToken    = "demo-token"
	ValidSubject  = "student"
)

type AuthService struct{}

func New() *AuthService {
	return &AuthService{}
}

func (s *AuthService) Login(username, password string) (string, bool) {
	if username == ValidUsername && password == ValidPassword {
		return ValidToken, true
	}
	return "", false
}

func (s *AuthService) Verify(token string) (subject string, valid bool) {
	if token == ValidToken {
		return ValidSubject, true
	}
	return "", false
}
