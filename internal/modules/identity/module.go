package identity

type IdentityService interface {
	HelloAuth() string
}

type IdentityServiceImpl struct {
}

func (*IdentityServiceImpl) HelloAuth() string {
	return "Hello from the Auth/Identity service!"
}

func NewIdentityService() *IdentityServiceImpl {
	return &IdentityServiceImpl{}
}
