package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"flag"
	"fmt"
	"io"
	"os"
)

// Ключ шифрования (должен совпадать с ключом в основном проекте)
var encryptionKey = []byte("0123456789abcdef0123456789abcdef") // 32 байта

// SetEncryptionKey устанавливает ключ шифрования
func SetEncryptionKey(key string) error {
	// Проверяем длину ключа
	if len(key) != 32 {
		return fmt.Errorf("ключ шифрования должен быть длиной 32 байта, получено %d байт", len(key))
	}
	encryptionKey = []byte(key)
	return nil
}

// DecryptPassword дешифрует пароль
func DecryptPassword(encryptedText string) (string, error) {
	// Убираем префикс "enc:" если есть
	if len(encryptedText) > 4 && encryptedText[:4] == "enc:" {
		encryptedText = encryptedText[4:]
	}

	data, err := base64.StdEncoding.DecodeString(encryptedText)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("неверная длина зашифрованных данных")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// EncryptPassword шифрует пароль
func EncryptPassword(plainText string) (string, error) {
	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plainText), nil)
	return "enc:" + base64.StdEncoding.EncodeToString(ciphertext), nil
}

func main() {
	// Определяем флаги
	encode := flag.Bool("encode", false, "Закодировать пароль")
	decode := flag.Bool("decode", false, "Декодировать пароль")
	key := flag.String("key", "", "Ключ шифрования (32 байта)")
	flag.Parse()

	// Проверяем ключ
	if *key != "" {
		if err := SetEncryptionKey(*key); err != nil {
			fmt.Printf("Ошибка установки ключа: %v\n", err)
			os.Exit(1)
		}
	} else {
		// Проверяем ключ по умолчанию
		if len(encryptionKey) != 32 {
			fmt.Printf("Ошибка: ключ по умолчанию должен быть 32 байта, текущий размер: %d байт\n", len(encryptionKey))
			fmt.Printf("Ключ по умолчанию: '%s'\n", string(encryptionKey))
			fmt.Println("Используйте флаг -key для указания своего ключа длиной 32 байта")
			os.Exit(1)
		}
	}

	// Проверяем что выбран только один режим
	if (*encode && *decode) || (!*encode && !*decode) {
		fmt.Println("Использование:")
		fmt.Println("  Кодирование: crypto_password -encode <пароль>")
		fmt.Println("  Декодирование: crypto_password -decode <зашифрованный_пароль>")
		fmt.Println("  Своим ключом: crypto_password -key '32-char-key' -encode <пароль>")
		fmt.Println("\nПример:")
		fmt.Println("  crypto_password -encode \"mysecret\"")
		fmt.Println("  crypto_password -decode \"enc:U2FsdGVkX1+WvFJzW1kQy8K5t6M8V9p7R2XbL3aN4cO0=\"")
		fmt.Println("  crypto_password -key \"32-byte-long-key-1234567890123456\" -encode \"password\"")
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Получаем пароль из аргументов
	args := flag.Args()
	if len(args) < 1 {
		fmt.Println("Ошибка: не указан пароль")
		os.Exit(1)
	}
	password := args[0]

	if *encode {
		// Режим кодирования
		encrypted, err := EncryptPassword(password)
		if err != nil {
			fmt.Printf("Ошибка кодирования: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Зашифрованный пароль: %s\n", encrypted)
		fmt.Printf("\nДля использования в config_historian.json:\n")
		fmt.Printf("\"password\": \"%s\"\n", encrypted)

	} else if *decode {
		// Режим декодирования
		decrypted, err := DecryptPassword(password)
		if err != nil {
			fmt.Printf("Ошибка декодирования: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Расшифрованный пароль: %s\n", decrypted)
	}
}

/*
1. Кодирование пароля:
cd E:\!!!VMWARE\VM_GO\windows\project\go-server\server-system\doc\script

go run crypto_password.go -encode "root"

2. Декодирование пароля:
go run crypto_password.go -decode "enc:U2FsdGVkX1+WvFJzW1kQy8K5t6M8V9p7R2XbL3aN4cO0="

3. С собственным ключом:
go run crypto_password.go -key "my-32-byte-long-encryption-key-123" -encode "password123"

Компиляция в исполняемый файл:
go build -o crypto_password crypto_password.go

После компиляции можно использовать:
./crypto_password -encode "my_password"
./crypto_password -decode "enc:U2FsdGVkX1+WvFJzW1kQy8K5t6M8V9p7R2XbL3aN4cO0="
*/
