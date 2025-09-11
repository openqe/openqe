package auth

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// GenerateHtpasswdBcrypt replicates `htpasswd -Bbn username password`
func GenerateHtpasswdBcrypt(username, password string) (string, error) {
	// Apache htpasswd uses cost=5 for bcrypt (-B)
	const bcryptCost = 5

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return "", err
	}
	// Apache uses $2y$ prefix instead of Go's $2a$
	// Replace it to match htpasswd output
	bcryptHash := string(hash)
	if bcryptHash[:4] == "$2a$" {
		bcryptHash = "$2y$" + bcryptHash[4:]
	}
	return fmt.Sprintf("%s:%s", username, bcryptHash), nil
}
