package service

type IdentityService interface {
	RegisterUser() (int, error)
}

type identityService struct {
}

func New() *identityService {
	return &identityService{}
}

func (s *identityService) RegisterUser() (int, error) {
	return 1, nil
}
