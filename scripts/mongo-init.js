// MongoDB initialization script for GoTask Management
// This script sets up the initial database structure and sample data

// Switch to the gotask database
db = db.getSiblingDB('gotask');

// Create a user for the application
db.createUser({
  user: 'gotask_user',
  pwd: 'gotask_password',
  roles: [
    {
      role: 'readWrite',
      db: 'gotask'
    }
  ]
});

// Create the tasks collection with validation schema
db.createCollection('tasks', {
  validator: {
    $jsonSchema: {
      bsonType: 'object',
      required: ['id', 'title', 'done', 'created_at'],
      properties: {
        id: {
          bsonType: 'string',
          description: 'Task ID must be a string and is required'
        },
        title: {
          bsonType: 'string',
          description: 'Task title must be a string and is required'
        },
        done: {
          bsonType: 'bool',
          description: 'Task completion status must be a boolean and is required'
        },
        created_at: {
          bsonType: 'date',
          description: 'Creation timestamp must be a date and is required'
        },
        due_date: {
          bsonType: ['date', 'null'],
          description: 'Due date must be a date or null'
        }
      }
    }
  },
  validationLevel: 'moderate',
  validationAction: 'warn'
});

// Create indexes for better performance
db.tasks.createIndex({ 'id': 1 }, { unique: true });
db.tasks.createIndex({ 'created_at': -1 });
db.tasks.createIndex({ 'due_date': 1 });
db.tasks.createIndex({ 'done': 1 });
db.tasks.createIndex({ 'done': 1, 'due_date': 1 });

// Create a text index for full-text search on title
db.tasks.createIndex({ 'title': 'text' });

// Insert sample data (optional)
// Uncomment the following lines if you want sample data

/*
db.tasks.insertMany([
  {
    id: 'sample-1',
    title: 'Welcome to GoTask Management! 🚀',
    done: false,
    created_at: new Date(),
    due_date: new Date(Date.now() + 7 * 24 * 60 * 60 * 1000) // 7 days from now
  },
  {
    id: 'sample-2',
    title: 'Configure your storage backend',
    done: false,
    created_at: new Date(),
    due_date: new Date(Date.now() + 3 * 24 * 60 * 60 * 1000) // 3 days from now
  },
  {
    id: 'sample-3',
    title: 'Explore the API endpoints',
    done: false,
    created_at: new Date(),
    due_date: new Date(Date.now() + 5 * 24 * 60 * 60 * 1000) // 5 days from now
  },
  {
    id: 'sample-4',
    title: 'Set up monitoring and logging',
    done: false,
    created_at: new Date(),
    due_date: new Date(Date.now() + 14 * 24 * 60 * 60 * 1000) // 14 days from now
  },
  {
    id: 'sample-5',
    title: 'Deploy to production',
    done: false,
    created_at: new Date(),
    due_date: new Date(Date.now() + 30 * 24 * 60 * 60 * 1000) // 30 days from now
  },
  {
    id: 'sample-6',
    title: 'Test with UTF-8: 你好世界 ñáéíóú',
    done: false,
    created_at: new Date(),
    due_date: new Date(Date.now() + 1 * 24 * 60 * 60 * 1000) // 1 day from now
  }
]);
*/

// Create a view for overdue tasks
db.createView('overdue_tasks', 'tasks', [
  {
    $match: {
      done: false,
      due_date: { $lt: new Date() }
    }
  },
  {
    $addFields: {
      days_overdue: {
        $divide: [
          { $subtract: [new Date(), '$due_date'] },
          1000 * 60 * 60 * 24
        ]
      }
    }
  },
  {
    $sort: { due_date: 1 }
  }
]);

// Create a view for upcoming tasks (due in next 7 days)
db.createView('upcoming_tasks', 'tasks', [
  {
    $match: {
      done: false,
      due_date: {
        $gte: new Date(),
        $lte: new Date(Date.now() + 7 * 24 * 60 * 60 * 1000)
      }
    }
  },
  {
    $addFields: {
      days_until_due: {
        $divide: [
          { $subtract: ['$due_date', new Date()] },
          1000 * 60 * 60 * 24
        ]
      }
    }
  },
  {
    $sort: { due_date: 1 }
  }
]);

// Create aggregation functions as stored procedures (using MongoDB's stored JavaScript)

// Function to get task statistics
db.system.js.save({
  _id: 'getTaskStatistics',
  value: function() {
    return db.tasks.aggregate([
      {
        $group: {
          _id: null,
          total_tasks: { $sum: 1 },
          completed_tasks: {
            $sum: { $cond: [{ $eq: ['$done', true] }, 1, 0] }
          },
          pending_tasks: {
            $sum: { $cond: [{ $eq: ['$done', false] }, 1, 0] }
          },
          overdue_tasks: {
            $sum: {
              $cond: [
                {
                  $and: [
                    { $eq: ['$done', false] },
                    { $lt: ['$due_date', new Date()] }
                  ]
                },
                1,
                0
              ]
            }
          },
          due_this_week: {
            $sum: {
              $cond: [
                {
                  $and: [
                    { $ne: ['$due_date', null] },
                    { $gte: ['$due_date', new Date()] },
                    { $lte: ['$due_date', new Date(Date.now() + 7 * 24 * 60 * 60 * 1000)] }
                  ]
                },
                1,
                0
              ]
            }
          }
        }
      }
    ]).toArray()[0];
  }
});

// Function to clean up old completed tasks
db.system.js.save({
  _id: 'cleanupOldCompletedTasks',
  value: function(daysOld) {
    if (!daysOld) daysOld = 30;
    
    const cutoffDate = new Date(Date.now() - daysOld * 24 * 60 * 60 * 1000);
    
    const result = db.tasks.deleteMany({
      done: true,
      created_at: { $lt: cutoffDate }
    });
    
    return {
      deleted_count: result.deletedCount,
      cutoff_date: cutoffDate
    };
  }
});

// Function to get completion rate
db.system.js.save({
  _id: 'getCompletionRate',
  value: function() {
    const stats = db.tasks.aggregate([
      {
        $group: {
          _id: null,
          total: { $sum: 1 },
          completed: {
            $sum: { $cond: [{ $eq: ['$done', true] }, 1, 0] }
          }
        }
      }
    ]).toArray()[0];
    
    if (!stats || stats.total === 0) {
      return 0;
    }
    
    return (stats.completed / stats.total) * 100;
  }
});

// Create a capped collection for audit logs (optional)
db.createCollection('task_audit_log', {
  capped: true,
  size: 10485760, // 10MB
  max: 10000      // Maximum 10,000 documents
});

// Create an index on the audit log timestamp
db.task_audit_log.createIndex({ 'timestamp': -1 });

print('MongoDB database initialized successfully for GoTask Management');
print('Created collections: tasks, overdue_tasks (view), upcoming_tasks (view), task_audit_log');
print('Created indexes on: id (unique), created_at, due_date, done, title (text)');
print('Created stored functions: getTaskStatistics, cleanupOldCompletedTasks, getCompletionRate');
print('Created user: gotask_user with readWrite permissions');
