-- Create users table
CREATE TABLE IF NOT EXISTS users (
    id INT AUTO_INCREMENT PRIMARY KEY,
    username VARCHAR(255) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_username (username)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Alter items table to add audit fields
ALTER TABLE items
ADD COLUMN created_by INT NULL,
ADD COLUMN updated_by INT NULL,
ADD CONSTRAINT fk_items_created_by FOREIGN KEY (created_by) REFERENCES users(id),
ADD CONSTRAINT fk_items_updated_by FOREIGN KEY (updated_by) REFERENCES users(id);
