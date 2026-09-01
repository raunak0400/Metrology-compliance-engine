// Password hashing with Argon2id (matches the original Node service's use of
// the `argon2` package with its default argon2id settings).
package auth

import "github.com/alexedwards/argon2id"

func HashPassword(plain string) (string, error) {
	return argon2id.CreateHash(plain, argon2id.DefaultParams)
}

func VerifyPassword(hash, plain string) (bool, error) {
	match, _, err := argon2id.CheckHash(plain, hash)
	if err != nil {
		return false, err
	}
	return match, nil
}
