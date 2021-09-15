package auth

import "net/http"


func (au *AuthHandler) VerifyMail(w http.ResponseWriter, r *http.Request) {}

func (au *AuthHandler) VerifyPasswordReset(w http.ResponseWriter, r *http.Request){}

func (au *AuthHandler) GeneratePassResetCode(w http.ResponseWriter, r *http.Request){}