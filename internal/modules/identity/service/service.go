package service

type IdentityService interface {
	InterfaceTest() string
}

type identityService struct {
}

func New() *identityService {
	return &identityService{}
}

func (service *identityService) InterfaceTest() string {
	return "Hello from the auth service, this string is going thru a lot of places"
}
