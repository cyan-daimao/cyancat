// Package keychain 提供密码安全存储
//
// 存储优先级：
//  1. 系统 Keychain（macOS Keychain / Windows Credential Manager / Linux Secret Service）
//  2. 兜底方案：AES-256 加密后存入本地 SQLite
package keychain

import (
	"errors"
	"os"

	"cyancat/internal/infra/crypto"
)

var masterKey []byte

// Init 初始化密钥管理器
// masterKey 由用户主密码通过 PBKDF2 派生（V2.0 前使用固定 key 文件兜底）
func Init() error {
	// V1.0 从环境变量或本地文件读取 master key
	keyHex := os.Getenv("CYANCAT_MASTER_KEY")
	if keyHex != "" {
		// 期望 64 位 hex 字符串 = 32 字节
		key := make([]byte, 32)
		if n, err := decodeHex(key, keyHex); err != nil || n != 32 {
			return errors.New("keychain: CYANCAT_MASTER_KEY must be 64 hex chars")
		}
		masterKey = key
		return nil
	}

	// V1.0 使用固定 key 文件（~/.cyancat/master.key），不适用于生产
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	keyPath := home + "/.cyancat/master.key"
	keyData, err := os.ReadFile(keyPath)
	if err != nil {
		return err
	}

	if len(keyData) != 32 {
		return errors.New("keychain: master.key must be 32 bytes")
	}
	masterKey = keyData
	return nil
}

// Store 加密存储密码
func Store(service, account, password string) error {
	// V1.0 先实现 AES 兜底方案
	// TODO V2.0: 接入系统 Keychain（zalando/go-keyring）
	return storeLocal(service, account, password)
}

// Get 读取已存储的密码
func Get(service, account string) (string, error) {
	return getLocal(service, account)
}

// Delete 删除已存储的密码
func Delete(service, account string) error {
	return deleteLocal(service, account)
}

// --- 本地 AES 兜底方案 ---

func storeLocal(service, account, password string) error {
	if masterKey == nil {
		return errors.New("keychain: not initialized")
	}

	data := []byte(password)
	encrypted, err := crypto.Encrypt(data, masterKey)
	if err != nil {
		return err
	}

	// 由调用方负责持久化加密后的字符串
	// 这里存储到内存，caller 自行写入 SQLite
	localCache[localKey(service, account)] = encrypted
	return nil
}

func getLocal(service, account string) (string, error) {
	if masterKey == nil {
		return "", errors.New("keychain: not initialized")
	}

	encrypted, ok := localCache[localKey(service, account)]
	if !ok {
		return "", errors.New("keychain: credential not found")
	}

	decrypted, err := crypto.Decrypt(encrypted, masterKey)
	if err != nil {
		return "", err
	}

	return string(decrypted), nil
}

func deleteLocal(service, account string) error {
	delete(localCache, localKey(service, account))
	return nil
}

var localCache = make(map[string]string)

func localKey(service, account string) string {
	return service + ":" + account
}

func decodeHex(dst []byte, src string) (int, error) {
	if len(src)%2 != 0 {
		return 0, errors.New("keychain: invalid hex string")
	}
	for i := 0; i < len(src) && i/2 < len(dst); i += 2 {
		high := hexVal(src[i])
		low := hexVal(src[i+1])
		if high < 0 || low < 0 {
			return 0, errors.New("keychain: invalid hex char")
		}
		dst[i/2] = byte(high<<4 | low)
	}
	return len(src) / 2, nil
}

func hexVal(c byte) int {
	switch {
	case '0' <= c && c <= '9':
		return int(c - '0')
	case 'a' <= c && c <= 'f':
		return int(c - 'a' + 10)
	case 'A' <= c && c <= 'F':
		return int(c - 'A' + 10)
	default:
		return -1
	}
}