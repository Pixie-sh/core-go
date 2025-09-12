package validation

import "net/mail"

func IsValidEmail(email string) (bool, error) {
	_, err := mail.ParseAddress(email)
	return err == nil, err
}
