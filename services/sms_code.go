package services

import "ModuleForTestTask/repositories"

type SMSService struct {
	SMSCodeRepository repositories.SMSCodeRepository
}

func NewSMSService(smsCodeRepository repositories.SMSCodeRepository) *SMSService {
	return &SMSService{SMSCodeRepository: smsCodeRepository}
}
