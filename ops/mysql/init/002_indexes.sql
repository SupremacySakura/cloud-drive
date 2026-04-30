-- 创建用于加速查询的索引
CREATE INDEX idx_files_hash_user ON file_models(file_hash, user_id);
CREATE INDEX idx_files_parent_user ON file_models(parent_id, user_id);
CREATE INDEX idx_folders_parent_user ON folder_models(parent_id, user_id);
