
CREATE TABLE IF NOT EXISTS items (
    id SERIAL PRIMARY KEY,
    code VARCHAR(50) UNIQUE NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    price BIGINT NOT NULL,
    stock INTEGER NOT NULL,
    status VARCHAR(10) NOT NULL DEFAULT 'ACTIVE',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CHECK (status IN ('ACTIVE', 'INACTIVE'))
);


CREATE INDEX IF NOT EXISTS idx_items_status ON items(status);


CREATE INDEX IF NOT EXISTS idx_items_code ON items(code);


CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_items_updated_at
BEFORE UPDATE ON items
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

CREATE OR REPLACE FUNCTION update_item_status()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.stock <= 0 THEN
        NEW.status = 'INACTIVE';
    ELSE
        NEW.status = 'ACTIVE';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_items_status_trigger
BEFORE INSERT OR UPDATE OF stock ON items
FOR EACH ROW
EXECUTE FUNCTION update_item_status();


DROP TRIGGER IF EXISTS update_items_status_trigger ON items;
DROP FUNCTION IF EXISTS update_item_status();
DROP TRIGGER IF EXISTS update_items_updated_at ON items;
DROP FUNCTION IF EXISTS update_updated_at_column();
DROP INDEX IF EXISTS idx_items_code;
DROP INDEX IF EXISTS idx_items_status;
DROP TABLE IF EXISTS items;
