# 加密密钥迁移脚本

这个目录包含用于迁移加密密钥的工具脚本。

## 使用方法

1. **停止 GPT-Load 服务**
2. **备份数据库**
3. **修改 `migrate_encryption_key.go` 中的配置**：
   ```go
   const (
       DATABASE_DSN = "./data/gpt-load.db"  // 数据库连接字符串
       OLD_ENCRYPTION_KEY = "old-key"       // 当前加密密钥
       NEW_ENCRYPTION_KEY = "new-key"       // 新的加密密钥
   )
   ```
4. **运行迁移**：
   ```bash
   cd scripts
   go run migrate_encryption_key.go
   ```
5. **更新 .env 文件中的 ENCRYPTION_KEY**
6. **重启服务**

## 重要提醒

- ⚠️ **迁移前务必备份数据库**
- ⚠️ **确保服务已完全停止**
- ⚠️ **在生产环境使用前，请先在测试环境验证**

## 支持的场景

- **启用加密**：`OLD_ENCRYPTION_KEY = ""`，提供 `NEW_ENCRYPTION_KEY`
- **禁用加密**：提供当前密钥，`NEW_ENCRYPTION_KEY = ""`
- **更改密钥**：提供旧密钥和新密钥
