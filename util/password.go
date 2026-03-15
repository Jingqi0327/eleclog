package util

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// 将密码进行加密
func HashPassword(password string) (string, error) {
	//传入byte类型的密码和cost参数，生成加密后的密码
	hashPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("生成密码失败: %v", err)
	}
	return string(hashPassword), nil
}

// 验证密码是否正确
func CheckPassword(password, hashPassword string) error {
	//将加密后的密码和用户输入的密码进行比较
	err := bcrypt.CompareHashAndPassword([]byte(hashPassword), []byte(password))
	if err != nil {
		return err
	}
	return nil
}
