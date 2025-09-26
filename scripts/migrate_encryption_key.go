package main

import (
	"fmt"
	"log"

	"gpt-load/internal/encryption"
	"gpt-load/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// 加密密钥迁移脚本
// 使用方法：
// 1. 停止 GPT-Load 服务
// 2. 备份数据库
// 3. 修改下面的配置
// 4. 运行：go run migrate_encryption_key.go
// 5. 更新 .env 中的 ENCRYPTION_KEY
// 6. 重启服务

const (
	// 数据库连接字符串 - 请根据实际情况修改
	DATABASE_DSN = "./data/gpt-load.db"

	// 旧的加密密钥（当前正在使用的）
	OLD_ENCRYPTION_KEY = ""

	// 新的加密密钥（要切换到的）
	NEW_ENCRYPTION_KEY = "your-new-32-char-secret-key-here"
)

func main() {
	if OLD_ENCRYPTION_KEY == "" && NEW_ENCRYPTION_KEY == "" {
		log.Fatal("请设置 OLD_ENCRYPTION_KEY 和 NEW_ENCRYPTION_KEY")
	}

	// 连接数据库
	db, err := connectDB(DATABASE_DSN)
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}

	// 创建加密服务
	oldService, err := encryption.NewService(OLD_ENCRYPTION_KEY)
	if err != nil {
		log.Fatalf("创建旧加密服务失败: %v", err)
	}

	newService, err := encryption.NewService(NEW_ENCRYPTION_KEY)
	if err != nil {
		log.Fatalf("创建新加密服务失败: %v", err)
	}

	// 获取所有 API 密钥
	var keys []models.APIKey
	if err := db.Find(&keys).Error; err != nil {
		log.Fatalf("查询 API 密钥失败: %v", err)
	}

	fmt.Printf("找到 %d 个 API 密钥需要迁移\n", len(keys))

	// 迁移每个密钥
	for i, key := range keys {
		fmt.Printf("迁移密钥 %d/%d (ID: %d)...", i+1, len(keys), key.ID)

		// 解密旧密钥
		plaintext, err := oldService.Decrypt(key.KeyValue)
		if err != nil {
			// 如果解密失败，可能是未加密的数据
			plaintext = key.KeyValue
			fmt.Printf(" (未加密)")
		}

		// 使用新密钥加密
		encrypted, err := newService.Encrypt(plaintext)
		if err != nil {
			log.Fatalf("加密密钥失败: %v", err)
		}

		// 计算新的哈希
		newHash := newService.Hash(plaintext)

		// 更新数据库
		if err := db.Model(&key).Updates(map[string]interface{}{
			"key_value": encrypted,
			"key_hash":  newHash,
		}).Error; err != nil {
			log.Fatalf("更新密钥失败: %v", err)
		}

		fmt.Println(" ✓")
	}

	fmt.Printf("迁移完成！请更新 .env 文件中的 ENCRYPTION_KEY 为: %s\n", NEW_ENCRYPTION_KEY)
}

func connectDB(dsn string) (*gorm.DB, error) {
	// 尝试 SQLite
	if db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{}); err == nil {
		return db, nil
	}

	// 尝试 PostgreSQL
	if db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{}); err == nil {
		return db, nil
	}

	return nil, fmt.Errorf("无法连接数据库")
}
