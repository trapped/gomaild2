package db

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"encoding/base64"
	log "github.com/sirupsen/logrus"
	config "github.com/spf13/viper"
	. "github.com/trapped/gomaild2/structs"
)

var (
	disablePWEncryption bool
	decodedKey          []byte
	openSSLSaltHeader   string = "Salted_"
)

func initEncryption() {
	WaitConfig("config.loaded")
	key := config.GetString("pw_encryption")
	decoded_key, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		log.Fatal(err)
	}
	if len(decoded_key) == 0 {
		disablePWEncryption = true
		log.Warn("Password encryption disabled")
		return
	} else if len(decoded_key) != 32 {
		log.Fatal("Password encryption key size must be 32 bytes (is: ", len(decoded_key), " bytes)")
	}
	decodedKey = decoded_key
	config.Set("encryption.loaded", true)
}

type openSSLCreds struct {
	key []byte
	iv  []byte
}

// Decrypt string that was encrypted using OpenSSL and AES-256-CBC
func __decryptString(key []byte, enc string) string {
	data, err := base64.StdEncoding.DecodeString(enc)
	if err != nil {
		log.Fatal(err)
	}
	saltHeader := data[:aes.BlockSize]
	if string(saltHeader[:7]) != openSSLSaltHeader {
		log.Fatal("String doesn't appear to have been encrypted with OpenSSL")
	}
	salt := saltHeader[8:]
	creds := __extractOpenSSLCreds(key, salt)
	if err != nil {
		log.Fatal(err)
	}
	return string(__decrypt(creds.key, creds.iv, data))
}

func __decrypt(key, iv, data []byte) []byte {
	if len(data) == 0 || len(data)%aes.BlockSize != 0 {
		log.Fatal("Bad blocksize(%v), aes.BlockSize = %v\n", len(data), aes.BlockSize)
	}
	c, err := aes.NewCipher(key)
	if err != nil {
		log.Fatal(err)
	}
	cbc := cipher.NewCBCDecrypter(c, iv)
	cbc.CryptBlocks(data[aes.BlockSize:], data[aes.BlockSize:])
	out := pkcs7Unpad(data[aes.BlockSize:], aes.BlockSize)
	return out
}

// openSSLEvpBytesToKey follows the OpenSSL (undocumented?) convention for extracting the key and IV from passphrase.
// It uses the EVP_BytesToKey() method which is basically:
// D_i = HASH^count(D_(i-1) || password || salt) where || denotes concatentaion, until there are sufficient bytes available
// 48 bytes since we're expecting to handle AES-256, 32bytes for a key and 16bytes for the IV
func __extractOpenSSLCreds(password, salt []byte) openSSLCreds {
	m := make([]byte, 48)
	prev := []byte{}
	for i := 0; i < 3; i++ {
		prev = __hash(prev, password, salt)
		copy(m[i*16:], prev)
	}
	return openSSLCreds{key: m[:32], iv: m[32:]}
}

func __hash(prev, password, salt []byte) []byte {
	a := make([]byte, len(prev)+len(password)+len(salt))
	copy(a, prev)
	copy(a[len(prev):], password)
	copy(a[len(prev)+len(password):], salt)
	return md5sum(a)
}

func md5sum(data []byte) []byte {
	h := md5.New()
	h.Write(data)
	return h.Sum(nil)
}

// pkcs7Unpad returns slice of the original data without padding.
func pkcs7Unpad(data []byte, blocklen int) []byte {
	if blocklen <= 0 {
		log.Fatal("Invalid blocklen %d", blocklen)
	}
	if len(data)%blocklen != 0 || len(data) == 0 {
		log.Fatal("Invalid data len %d", len(data))
	}
	padlen := int(data[len(data)-1])
	if padlen > blocklen || padlen == 0 {
		log.Fatal("Invalid padding")
	}
	pad := data[len(data)-padlen:]
	for i := 0; i < padlen; i++ {
		if pad[i] != byte(padlen) {
			log.Fatal("Invalid padding")
		}
	}
	return data[:len(data)-padlen]
}

func decryptPassword(encrypted string) string {
	WaitConfig("encryption.loaded")
	if !disablePWEncryption {
		return __decryptString(decodedKey, encrypted)
	}
	return encrypted
}
