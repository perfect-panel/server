-- Only add the columns to `servers` when they do not already exist

-- Add longitude
SET @col_exists := (
    SELECT COUNT(*)
    FROM INFORMATION_SCHEMA.COLUMNS
    WHERE TABLE_SCHEMA = DATABASE()
      AND TABLE_NAME = 'servers'
      AND COLUMN_NAME = 'longitude'
);

SET @query := IF(
    @col_exists = 0,
    'ALTER TABLE `servers` ADD COLUMN `longitude` VARCHAR(255) DEFAULT '''' COMMENT ''longitude'';',
    'SELECT "Column `longitude` already exists"'
);

PREPARE stmt FROM @query;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- Add latitude
SET @col_exists := (
    SELECT COUNT(*)
    FROM INFORMATION_SCHEMA.COLUMNS
    WHERE TABLE_SCHEMA = DATABASE()
      AND TABLE_NAME = 'servers'
      AND COLUMN_NAME = 'latitude'
);

SET @query := IF(
    @col_exists = 0,
    'ALTER TABLE `servers` ADD COLUMN `latitude` VARCHAR(255) DEFAULT '''' COMMENT ''latitude'';',
    'SELECT "Column `latitude` already exists"'
);

PREPARE stmt FROM @query;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- Add longitude_center
SET @col_exists := (
    SELECT COUNT(*)
    FROM INFORMATION_SCHEMA.COLUMNS
    WHERE TABLE_SCHEMA = DATABASE()
      AND TABLE_NAME = 'servers'
      AND COLUMN_NAME = 'longitude_center'
);

SET @query := IF(
    @col_exists = 0,
    'ALTER TABLE `servers` ADD COLUMN `longitude_center` VARCHAR(255) DEFAULT '''' COMMENT ''longitude center'';',
    'SELECT "Column `longitude_center` already exists"'
);

PREPARE stmt FROM @query;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- Add latitude_center
SET @col_exists := (
    SELECT COUNT(*)
    FROM INFORMATION_SCHEMA.COLUMNS
    WHERE TABLE_SCHEMA = DATABASE()
      AND TABLE_NAME = 'servers'
      AND COLUMN_NAME = 'latitude_center'
);

SET @query := IF(
    @col_exists = 0,
    'ALTER TABLE `servers` ADD COLUMN `latitude_center` VARCHAR(255) DEFAULT '''' COMMENT ''latitude center'';',
    'SELECT "Column `latitude_center` already exists"'
);

PREPARE stmt FROM @query;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

