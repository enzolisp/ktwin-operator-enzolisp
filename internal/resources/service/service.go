package service

type TwinService interface {
	CreateService()
	DeleteService()
}

type twinService struct{}

func (*twinService) CreateService() {}

func (*twinService) DeleteService() {}
