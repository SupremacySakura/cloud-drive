CREATE DATABASE IF NOT EXISTS `cloud-drive` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE `cloud-drive`;

CREATE TABLE IF NOT EXISTS `bootstrap_records` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `tag` VARCHAR(64) NOT NULL,
  `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_bootstrap_tag` (`tag`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

INSERT INTO `bootstrap_records` (`tag`)
VALUES ('mysql-initdb-loaded')
ON DUPLICATE KEY UPDATE `tag` = VALUES(`tag`);
