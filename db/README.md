# Database Migrations

This directory contains the database migration system for the EkklesiaHR project.

## Directory Structure

```
db/
├── migrations/           # Migration files
│   ├── 000001_create_schema_migrations.up.sql
│   ├── 000001_create_schema_migrations.down.sql
│   ├── 000002_create_users_table.up.sql
│   └── 000002_create_users_table.down.sql
└── README.md            # This file
```

## Migration Naming Convention

Migration files should follow this naming pattern:
```
NNNNNN_description.up.sql    # For applying the migration
NNNNNN_description.down.sql  # For rolling back the migration
```

Where:
- `NNNNNN` is a 6-digit version number (e.g., 000001, 000002)
- `description` is a snake_case description of what the migration does
- `.up.sql` contains the SQL to apply the migration
- `.down.sql` contains the SQL to rollback the migration

## Usage

### Apply all pending migrations
```bash
make migrate-up
```

### Rollback the last migration
```bash
make migrate-down
```

### Rollback all migrations
```bash
make migrate-down-all
```

### Check migration status
```bash
make migrate-status
```

### Create a new migration
```bash
make migrate-create MIGRATION_NAME="add_employee_table"
```

### Direct CLI usage
You can also use the migration CLI directly:

```bash
# Apply migrations
go run cmd/migrate/main.go -command=up

# Rollback last migration
go run cmd/migrate/main.go -command=down -steps=1

# Check status
go run cmd/migrate/main.go -command=status

# Create new migration
go run cmd/migrate/main.go -command=create -name="your_migration_name"
```

## Migration Tracking

The system uses a `schema_migrations` table to track which migrations have been applied:

- `version`: The migration version number
- `dirty`: Whether the migration is in a dirty state (failed during application)
- `applied_at`: When the migration was successfully applied

## Best Practices

1. **Always create both up and down migrations** - This allows for rollbacks
2. **Test migrations thoroughly** - Test both up and down migrations
3. **Make migrations idempotent** - Use `IF NOT EXISTS` and `IF EXISTS` clauses
4. **Keep migrations atomic** - Each migration should be a single logical unit
5. **Don't modify existing migrations** - Once applied to production, create a new migration instead
6. **Use descriptive names** - Make it clear what each migration does

## Example Migration

### Up Migration (000003_add_employee_table.up.sql)
```sql
CREATE TABLE IF NOT EXISTS employees (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    employee_id VARCHAR(50) UNIQUE NOT NULL,
    department_id UUID,
    hire_date DATE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_employees_user_id ON employees(user_id);
CREATE INDEX IF NOT EXISTS idx_employees_employee_id ON employees(employee_id);
```

### Down Migration (000003_add_employee_table.down.sql)
```sql
DROP INDEX IF EXISTS idx_employees_employee_id;
DROP INDEX IF EXISTS idx_employees_user_id;
DROP TABLE IF EXISTS employees;
```

## Troubleshooting

### Migration is marked as dirty
If a migration fails during application, it will be marked as "dirty". To resolve:

1. Fix the issue that caused the migration to fail
2. Manually clean up any partial changes if necessary
3. Update the migration state in the database:
   ```sql
   UPDATE schema_migrations SET dirty = FALSE WHERE version = 'XXXXXX';
   ```
4. Re-run the migration

### Migration rollback fails
If a rollback fails:

1. Check the down migration SQL for errors
2. Manually verify the database state
3. Fix any issues and retry the rollback
4. If necessary, manually clean up and update the migration tracking table