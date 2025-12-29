-- Create the database if it doesn't exist
CREATE DATABASE IF NOT EXISTS mailculator;
USE mailculator;

-- Emails table
CREATE TABLE IF NOT EXISTS emails (
    id CHAR(36) PRIMARY KEY,
    status ENUM(
        'ACCEPTED','INTAKING','READY','PROCESSING',
        'SENT','FAILED','INVALID',
        'CALLING-SENT-CALLBACK','CALLING-FAILED-CALLBACK',
        'SENT-ACKNOWLEDGED','FAILED-ACKNOWLEDGED'
    ) NOT NULL,
    eml_file_path VARCHAR(500),
    payload_file_path VARCHAR(500),
    reason TEXT,
    version INT NOT NULL DEFAULT 1,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    INDEX idx_status (status),
    INDEX idx_status_updated (status, updated_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- Email statuses history table
CREATE TABLE IF NOT EXISTS email_statuses (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    email_id CHAR(36) NOT NULL,
    status VARCHAR(50) NOT NULL,
    reason TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    INDEX idx_email_id (email_id),
    FOREIGN KEY (email_id) REFERENCES emails(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

