package crypto_utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"github.com/harshasavanth/bookstore_users-api/logger"
	"github.com/harshasavanth/utils-go/rest_errors"
	"io"
	"os"
)

func GetMd5(input string) string {
	hash := md5.New()
	defer hash.Reset()
	hash.Write([]byte(input))
	return hex.EncodeToString(hash.Sum(nil))
}

const (
	key = "key"
)

func Encrypt(stringToEncrypt string) (string, *rest_errors.RestErr) {

	//Since the key is in string, we need to convert decode it to bytes
	key, _ := hex.DecodeString(os.Getenv(key))
	plaintext := []byte(stringToEncrypt)

	//Create a new Cipher Block from the key
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", rest_errors.NewInvalidInputError("could not generate link")
	}

	//Create a new GCM - https://en.wikipedia.org/wiki/Galois/Counter_Mode
	//https://golang.org/pkg/crypto/cipher/#NewGCM
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", rest_errors.NewInvalidInputError("could not generate link")
	}

	//Create a nonce. Nonce should be from GCM
	nonce := make([]byte, aesGCM.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", rest_errors.NewInvalidInputError("could not generate link")
	}

	//Encrypt the data using aesGCM.Seal
	//Since we don't want to save the nonce somewhere else in this case, we add it as a prefix to the encrypted data. The first nonce argument in Seal is the prefix.
	ciphertext := aesGCM.Seal(nonce, nonce, plaintext, nil)
	return fmt.Sprintf("%x", ciphertext), nil
}

func Decrypt(encryptedString string) (string, *rest_errors.RestErr) {
	logger.Info(encryptedString)
	key, _ := hex.DecodeString(os.Getenv(key))
	enc, _ := hex.DecodeString(encryptedString)

	//Create a new Cipher Block from the key
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", rest_errors.NewInternalServerError("error while decrypting ")
	}

	//Create a new GCM
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", rest_errors.NewInternalServerError("error while decrypting ")
	}

	//Get the nonce size
	nonceSize := aesGCM.NonceSize()

	//Extract the nonce from the encrypted data
	nonce, ciphertext := enc[:nonceSize], enc[nonceSize:]

	//Decrypt the data
	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", rest_errors.NewBadRequestError("invalid link")
	}
	logger.Info(string(plaintext))
	return fmt.Sprintf("%s", plaintext), nil
}
