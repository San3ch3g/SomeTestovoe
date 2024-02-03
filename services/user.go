package services

import "ModuleForTestTask/repositories"

type UserService struct {
	UserRepository repositories.UserRepository
}

func NewUserService(userRepository repositories.UserRepository) *UserService {
	return &UserService{UserRepository: userRepository}
}
