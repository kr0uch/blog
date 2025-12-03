package hash

import "golang.org/x/crypto/bcrypt"

func HashString(str string) (string, error) {
	hashedStr, err := bcrypt.GenerateFromPassword([]byte(str), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedStr), nil
}

func CompareHashString(str string, hashedStr string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(hashedStr), []byte(str))
	return err == nil, err
}
