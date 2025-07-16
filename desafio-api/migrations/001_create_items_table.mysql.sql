CREATE TABLE IF NOT EXISTS items (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    code VARCHAR(255) NOT NULL UNIQUE,
    title VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    price BIGINT NOT NULL,
    stock INT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'INACTIVE',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_status (status),
    INDEX idx_code (code)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Trigger para atualizar o status baseado no estoque
DELIMITER //
CREATE TRIGGER before_item_update
BEFORE UPDATE ON items
FOR EACH ROW
BEGIN
    IF NEW.stock > 0 THEN
        SET NEW.status = 'ACTIVE';
    ELSE
        SET NEW.status = 'INACTIVE';
    END IF;
END //
DELIMITER ;
