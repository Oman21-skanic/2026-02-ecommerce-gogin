-- =============================================
-- ECOMMERCE DATABASE SETUP (PROFESSIONAL VERSION)
-- =============================================

-- 1. Create Database
CREATE DATABASE IF NOT EXISTS ecommerce
    CHARACTER SET utf8mb4
    COLLATE utf8mb4_unicode_ci;

USE ecommerce;



-- 2. Users Table
CREATE TABLE IF NOT EXISTS users (
    id CHAR(36) PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(20) NOT NULL DEFAULT 'user',
    email_verified BOOLEAN DEFAULT FALSE,
    created_at DATETIME NOT NULL,
    INDEX idx_email (email)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
-- 8. Email Verifications & Password Resets
CREATE TABLE IF NOT EXISTS email_verifications (
    token VARCHAR(255) PRIMARY KEY,
    user_id CHAR(36) NOT NULL,
    expires_at DATETIME NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS password_resets (
    token VARCHAR(255) PRIMARY KEY,
    user_id CHAR(36) NOT NULL,
    expires_at DATETIME NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 3. Products Table (Updated with Rich Data Fields)
CREATE TABLE IF NOT EXISTS products (
    id CHAR(36) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    category VARCHAR(100),            -- Filter kategori
    price_cents BIGINT NOT NULL,
    sku VARCHAR(100) NOT NULL,
    stock INT NOT NULL,
    thumbnail TEXT,                   -- URL gambar utama
    images TEXT,                      -- URL gambar tambahan (comma separated)
    rating DECIMAL(3,2) DEFAULT 4.5,  -- Social proof: Bintang
    review_count INT DEFAULT 0,       -- Social proof: Jumlah ulasan
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    INDEX idx_sku (sku),
    INDEX idx_name (name),
    INDEX idx_category (category)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 4. Carts Table
CREATE TABLE IF NOT EXISTS carts (
    user_id CHAR(36) PRIMARY KEY,
    updated_at DATETIME NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 5. Cart Items Table
CREATE TABLE IF NOT EXISTS cart_items (
    user_id CHAR(36) NOT NULL,
    product_id CHAR(36) NOT NULL,
    quantity INT NOT NULL,
    PRIMARY KEY (user_id, product_id),
    FOREIGN KEY (user_id) REFERENCES carts(user_id) ON DELETE CASCADE,
    FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 6. Orders Table
CREATE TABLE IF NOT EXISTS orders (
    id CHAR(36) PRIMARY KEY,
    user_id CHAR(36) NOT NULL,
    amount_cents BIGINT NOT NULL,
    status VARCHAR(20) NOT NULL, -- pending, paid, failed, cancelled
    payment_ref VARCHAR(255) NOT NULL,
    created_at DATETIME NOT NULL,
    INDEX idx_user_id (user_id),
    INDEX idx_created_at (created_at),
    FOREIGN KEY (user_id) REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 7. Order Items Table
CREATE TABLE IF NOT EXISTS order_items (
    order_id CHAR(36) NOT NULL,
    product_id CHAR(36) NOT NULL,
    quantity INT NOT NULL,
    price_cents BIGINT NOT NULL,
    PRIMARY KEY (order_id, product_id),
    FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE,
    FOREIGN KEY (product_id) REFERENCES products(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;


-- =============================================
-- SEED DATA (BIAR FE LANGSUNG CANTIK)
-- =============================================

INSERT INTO products (id, name, description, category, price_cents, sku, stock, thumbnail, images, rating, review_count, created_at, updated_at)
VALUES 
(UUID(), 'Keychron K2 V2 Mechanical Keyboard', 'Keyboard mekanik nirkabel dengan layout 75%, hot-swappable, dan pencahayaan RGB yang estetik.', 'Electronics', 1250000, 'K2V2-RGB', 50, 
'https://images.unsplash.com/photo-1595225476474-87563907a212', 
'https://images.unsplash.com/photo-1511467687858-23d96c32e4ae,https://images.unsplash.com/photo-1595044426077-d36d9236d54a', 
4.8, 156, NOW(), NOW()),

(UUID(), 'Sony WH-1000XM4 Noise Cancelling', 'Headphone premium dengan teknologi peredam bising terbaik di kelasnya dan kualitas audio resolusi tinggi.', 'Electronics', 3500000, 'SONY-XM4', 25, 
'https://images.unsplash.com/photo-1505740420928-5e560c06d30e', 
'https://images.unsplash.com/photo-1484704849700-f032a568e944,https://images.unsplash.com/photo-1524678606370-a47ad25cb82a', 
4.9, 89, NOW(), NOW()),

(UUID(), 'Minimalist Oversized Hoodie', 'Hoodie dengan potongan boxy yang nyaman, menggunakan bahan premium fleece 330gsm.', 'Fashion', 450000, 'HOOD-MIN', 100, 
'https://images.unsplash.com/photo-1556821840-3a63f95609a7', 
'https://images.unsplash.com/photo-1556821921-292a0763069c', 
4.7, 210, NOW(), NOW()),

(UUID(), 'Leather Daily Tote Bag', 'Tas kulit sintetis kualitas tinggi, tahan air, dan cocok untuk membawa laptop 14 inch.', 'Fashion', 299000, 'BAG-LTHR', 40, 
'https://images.unsplash.com/photo-1544816153-1629739556f8', 
'https://images.unsplash.com/photo-1591561954557-26941169b49e', 
4.6, 45, NOW(), NOW()),

(UUID(), 'Aesthetic Desk Lamp LED', 'Lampu meja minimalis dengan 3 mode warna cahaya, cocok untuk setup kerja produktif.', 'Home', 185000, 'LAMP-AST', 60, 
'https://images.unsplash.com/photo-1534073828943-f801091bb18c', 
'https://images.unsplash.com/photo-1507473885765-e6ed057f782c', 
4.5, 32, NOW(), NOW());