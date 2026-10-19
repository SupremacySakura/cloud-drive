-- 为应用表创建查询加速索引
-- 应用表由后端服务启动时自动迁移创建，首次初始化本脚本执行时表可能尚不存在；
-- 因此此处对「表不存在 / 索引已存在」两种情况均做容错，本脚本可在任意阶段重复执行。
-- 后端 internal/database/migrate.go 在 AutoMigrate 后补建这些索引。
DROP PROCEDURE IF EXISTS zhixu_ensure_index;
DELIMITER $$
CREATE PROCEDURE zhixu_ensure_index(
  IN p_table VARCHAR(64),
  IN p_index VARCHAR(64),
  IN p_columns VARCHAR(255)
)
BEGIN
  IF EXISTS (
    SELECT 1 FROM information_schema.tables
    WHERE table_schema = DATABASE() AND table_name = p_table
  ) AND NOT EXISTS (
    SELECT 1 FROM information_schema.statistics
    WHERE table_schema = DATABASE() AND table_name = p_table AND index_name = p_index
  ) THEN
    SET @ddl = CONCAT('CREATE INDEX ', p_index, ' ON ', p_table, ' (', p_columns, ')');
    PREPARE stmt FROM @ddl;
    EXECUTE stmt;
    DEALLOCATE PREPARE stmt;
  END IF;
END$$
DELIMITER ;

-- file_hash 为 TEXT 列，MySQL 索引需指定前缀长度；哈希值为定长摘要，64 字符前缀即可完整覆盖
CALL zhixu_ensure_index('file_models', 'idx_files_hash_user', 'file_hash(64), user_id');
CALL zhixu_ensure_index('file_models', 'idx_files_parent_user', 'folder_id, user_id');
CALL zhixu_ensure_index('folder_models', 'idx_folders_parent_user', 'parent_id, user_id');

DROP PROCEDURE IF EXISTS zhixu_ensure_index;
