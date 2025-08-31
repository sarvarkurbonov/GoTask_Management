-- MySQL initialization script for GoTask Management
-- This script sets up the initial database structure and sample data

-- Use the gotask database
USE gotask;

-- Set proper charset and collation for UTF-8 support
ALTER DATABASE gotask CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- Create the tasks table if it doesn't exist
-- Note: GORM will handle the actual table creation, but this ensures proper charset
CREATE TABLE IF NOT EXISTS tasks (
    id VARCHAR(255) PRIMARY KEY,
    title TEXT NOT NULL,
    done BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    due_date TIMESTAMP NULL,
    INDEX idx_tasks_due_date (due_date),
    INDEX idx_tasks_done (done),
    INDEX idx_tasks_created_at (created_at),
    INDEX idx_tasks_done_due_date (done, due_date)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Insert sample data (optional)
-- Uncomment the following lines if you want sample data

/*
INSERT IGNORE INTO tasks (id, title, done, created_at, due_date) VALUES
    ('sample-1', 'Welcome to GoTask Management! 🚀', FALSE, NOW(), DATE_ADD(NOW(), INTERVAL 7 DAY)),
    ('sample-2', 'Configure your storage backend', FALSE, NOW(), DATE_ADD(NOW(), INTERVAL 3 DAY)),
    ('sample-3', 'Explore the API endpoints', FALSE, NOW(), DATE_ADD(NOW(), INTERVAL 5 DAY)),
    ('sample-4', 'Set up monitoring and logging', FALSE, NOW(), DATE_ADD(NOW(), INTERVAL 14 DAY)),
    ('sample-5', 'Deploy to production', FALSE, NOW(), DATE_ADD(NOW(), INTERVAL 30 DAY)),
    ('sample-6', 'Test with UTF-8: 你好世界 ñáéíóú', FALSE, NOW(), DATE_ADD(NOW(), INTERVAL 1 DAY));
*/

-- Create a stored procedure to clean up old completed tasks
DELIMITER //
CREATE PROCEDURE CleanupOldCompletedTasks(IN days_old INT)
BEGIN
    DECLARE deleted_count INT DEFAULT 0;
    
    DELETE FROM tasks 
    WHERE done = TRUE 
    AND created_at < DATE_SUB(NOW(), INTERVAL days_old DAY);
    
    SET deleted_count = ROW_COUNT();
    SELECT deleted_count as deleted_tasks;
END //
DELIMITER ;

-- Create a stored procedure to get task statistics
DELIMITER //
CREATE PROCEDURE GetTaskStatistics()
BEGIN
    SELECT 
        COUNT(*) as total_tasks,
        SUM(CASE WHEN done = TRUE THEN 1 ELSE 0 END) as completed_tasks,
        SUM(CASE WHEN done = FALSE THEN 1 ELSE 0 END) as pending_tasks,
        SUM(CASE WHEN done = FALSE AND due_date < NOW() THEN 1 ELSE 0 END) as overdue_tasks,
        SUM(CASE WHEN due_date IS NOT NULL AND due_date BETWEEN NOW() AND DATE_ADD(NOW(), INTERVAL 7 DAY) THEN 1 ELSE 0 END) as due_this_week
    FROM tasks;
END //
DELIMITER ;

-- Create a view for overdue tasks
CREATE OR REPLACE VIEW overdue_tasks AS
SELECT 
    id,
    title,
    created_at,
    due_date,
    DATEDIFF(NOW(), due_date) as days_overdue
FROM tasks 
WHERE done = FALSE 
AND due_date < NOW()
ORDER BY due_date ASC;

-- Create a view for upcoming tasks (due in next 7 days)
CREATE OR REPLACE VIEW upcoming_tasks AS
SELECT 
    id,
    title,
    created_at,
    due_date,
    DATEDIFF(due_date, NOW()) as days_until_due
FROM tasks 
WHERE done = FALSE 
AND due_date BETWEEN NOW() AND DATE_ADD(NOW(), INTERVAL 7 DAY)
ORDER BY due_date ASC;

-- Create a function to calculate task completion rate
DELIMITER //
CREATE FUNCTION GetCompletionRate() 
RETURNS DECIMAL(5,2)
READS SQL DATA
DETERMINISTIC
BEGIN
    DECLARE total_count INT DEFAULT 0;
    DECLARE completed_count INT DEFAULT 0;
    DECLARE completion_rate DECIMAL(5,2) DEFAULT 0.00;
    
    SELECT COUNT(*) INTO total_count FROM tasks;
    
    IF total_count > 0 THEN
        SELECT COUNT(*) INTO completed_count FROM tasks WHERE done = TRUE;
        SET completion_rate = (completed_count / total_count) * 100;
    END IF;
    
    RETURN completion_rate;
END //
DELIMITER ;

-- Grant necessary permissions to gotask_user
GRANT ALL PRIVILEGES ON gotask.* TO 'gotask_user'@'%';
GRANT EXECUTE ON PROCEDURE gotask.CleanupOldCompletedTasks TO 'gotask_user'@'%';
GRANT EXECUTE ON PROCEDURE gotask.GetTaskStatistics TO 'gotask_user'@'%';
GRANT EXECUTE ON FUNCTION gotask.GetCompletionRate TO 'gotask_user'@'%';
GRANT SELECT ON gotask.overdue_tasks TO 'gotask_user'@'%';
GRANT SELECT ON gotask.upcoming_tasks TO 'gotask_user'@'%';

-- Flush privileges to ensure they take effect
FLUSH PRIVILEGES;

-- Log successful initialization
SELECT 'MySQL database initialized successfully for GoTask Management' as message;
