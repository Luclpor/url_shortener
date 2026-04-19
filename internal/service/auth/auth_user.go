package auth

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"github.com/Luclpor/url_shortener.git/internal/model"
	"github.com/google/uuid"
)

type UserRepository interface {
	AddNewUser(ctx context.Context) (*model.User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*model.User, error)
}

func generateRandom(size int) ([]byte, error) {
	// генерируем криптостойкие случайные байты в b
	b := make([]byte, size)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}

	return b, nil
}

type UserAuthentication interface {
	CreateEncryptedUser() (*model.User, string, error)
	DecryptUser(value string) (*model.User, error)
}

type UserAuth struct {
	aesgcm   cipher.AEAD
	userRepo UserRepository
}

func InitAuthService(key []byte, ur UserRepository) (UserAuthentication, error) {
	aesblock, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	aesgcm, err := cipher.NewGCM(aesblock)
	if err != nil {
		return nil, err
	}
	return &UserAuth{aesgcm, ur}, nil
}

func (ua *UserAuth) CreateEncryptedUser() (*model.User, string, error) {
	nonce, err := generateRandom(ua.aesgcm.NonceSize())
	if err != nil {
		return nil, "", err
	}

	user, err := ua.userRepo.AddNewUser(context.Background())
	if err != nil {
		return nil, "", err
	}

	ciphertext := ua.aesgcm.Seal(nil, nonce, user.ID[:], nil)
	data := append(nonce, ciphertext...)

	encryptedUserID := base64.RawURLEncoding.EncodeToString(data)

	return user, encryptedUserID, nil
}

func (ua *UserAuth) DecryptUser(value string) (*model.User, error) {
	data, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil {
		return nil, err
	}

	nonceSize := ua.aesgcm.NonceSize()
	if len(data) < nonceSize {
		return nil, fmt.Errorf("invalid encrypted user id")
	}

	nonce := data[:nonceSize]
	ciphertext := data[nonceSize:]

	plain, err := ua.aesgcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	id, err := uuid.FromBytes(plain)
	if err != nil {
		return nil, err
	}
	user, err := ua.userRepo.GetUserByID(context.Background(), id)
	if err != nil {
		return nil, err
	}
	return user, nil
}
