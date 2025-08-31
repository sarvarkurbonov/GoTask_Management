-- PostgreSQL initialization script for GoTask Management
-- This script sets up the initial database structure and sample data

-- Create the tasks table if it doesn't exist
-- Note: GORM will handle the actual table creation, but this ensures the database is ready
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create indexes for better performance
-- These will be created by GORM as well, but having them here ensures consistency
DO $$
BEGIN
    -- Check if the tasks table exists before creating indexes
    IF EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'tasks') THEN
        -- Create index on due_date for faster queries
        CREATE INDEX IF NOT EXISTS idx_tasks_due_date ON tasks(due_date);
        
        -- Create index on done status for filtering
        CREATE INDEX IF NOT EXISTS idx_tasks_done ON tasks(done);
        
        -- Create composite index for common queries
        CREATE INDEX IF NOT EXISTS idx_tasks_done_due_date ON tasks(done, due_date);
        
        -- Create index on created_at for ordering
        CREATE INDEX IF NOT EXISTS idx_tasks_created_at ON tasks(created_at);
    END IF;
END
$$;

-- Insert sample data (optional)
-- Uncomment the following lines if you want sample data

/*
INSERT INTO tasks (id, title, done, created_at, due_date) VALUES
    ('sample-1', 'Welcome to GoTask Management!', false, NOW(), NOW() + INTERVAL '7 days'),
    ('sample-2', 'Configure your storage backend', false, NOW(), NOW() + INTERVAL '3 days'),
    ('sample-3', 'Explore the API endpoints', false, NOW(), NOW() + INTERVAL '5 days'),
    ('sample-4', 'Set up monitoring and logging', false, NOW(), NOW() + INTERVAL '14 days'),
    ('sample-5', 'Deploy to production', false, NOW(), NOW() + INTERVAL '30 days')
ON CONFLICT (id) DO NOTHING;
*/

-- Create a function to clean up old completed tasks (optional)
CREATE OR REPLACE FUNCTION cleanup_old_completed_tasks(days_old INTEGER DEFAULT 30)
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    DELETE FROM tasks 
    WHERE done = true 
    AND created_at < NOW() - INTERVAL '1 day' * days_old;
    
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;

-- Create a function to get task statistics
CREATE OR REPLACE FUNCTION get_task_statistics()
RETURNS TABLE(
    total_tasks BIGINT,
    completed_tasks BIGINT,
    pending_tasks BIGINT,
    overdue_tasks BIGINT
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        COUNT(*) as total_tasks,
        COUNT(*) FILTER (WHERE done = true) as completed_tasks,
        COUNT(*) FILTER (WHERE done = false) as pending_tasks,
        COUNT(*) FILTER (WHERE done = false AND due_date < NOW()) as overdue_tasks
    FROM tasks;
END;
$$ LANGUAGE plpgsql;

-- Grant necessary permissions to the gotask_user
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO gotask_user;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO gotask_user;
GRANT ALL PRIVILEGES ON ALL FUNCTIONS IN SCHEMA public TO gotask_user;

-- Set default privileges for future objects
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO gotask_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO gotask_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON FUNCTIONS TO gotask_user;

-- Log successful initialization
DO $$
BEGIN
    RAISE NOTICE 'PostgreSQL database initialized successfully for GoTask Management';
END
$$;
