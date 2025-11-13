# Migration Validation Strategy

## Problem Analysis

### Root Cause
The original validation logic failed because it checked for River SDK tables (`river_job`, `river_leader`, `river_migration`) **before** they were created.

**Execution Order:**
```
1. migrations.Apply()
   ├─ Run 000001_init.sql → Create application tables ✅
   └─ validateSchema() → Check ALL tables including River ❌ FAIL
   
2. NewCoordinator() (never reached due to failure)
   └─ River SDK creates its tables ✅ (too late)
```

**Result:** Application crashes on startup with "critical tables missing: [river_job river_leader river_migration]"

---

## Solution Architecture

### Two-Level Validation System

#### Level 1: REQUIRED Tables (Hard Fail)
Tables that **MUST** exist after `000001_init.sql` runs:
- `instances`
- `webhook_configs`
- `event_outbox`
- `event_dlq`
- `media_metadata`
- `instance_event_sequence`
- `message_sequences`

**Behavior:** Fatal error if any missing → Database is in inconsistent state

#### Level 2: OPTIONAL Tables (Soft Warn)
Tables created by external systems (River SDK):
- `river_job`
- `river_leader`
- `river_migration`

**Behavior:** Log info message if missing → System continues normally

---

## Why This Works in All Environments

### Scenario 1: Fresh Database (First Run)
```
1. migrations.Apply() creates app tables
2. validateSchema() checks required tables → PASS ✅
3. validateSchema() checks optional tables → NOT FOUND (logs info, continues)
4. Returns SUCCESS
5. main.go continues → NewCoordinator() → River creates its tables
6. System fully operational
```

### Scenario 2: Existing Database
```
1. migrations.Apply() skips SQL (already applied)
2. validateSchema() checks required tables → PASS ✅
3. validateSchema() checks optional tables → FOUND ✅
4. Returns SUCCESS
5. System continues normally
```

### Scenario 3: Corrupted Database
```
1. migrations.Apply() may have run partially
2. validateSchema() checks required tables → FAIL ❌ (missing critical table)
3. Returns ERROR with clear message
4. System does not start (correct behavior - prevents data corruption)
```

### Scenario 4: CI/CD with Isolated Migrations
```
1. CI runs only migrations.Apply() for testing
2. Validates only application tables
3. Passes validation even without initializing entire system
4. No false negatives in CI pipeline
```

### Scenario 5: MessageQueue Disabled
```
1. migrations.Apply() validates app tables → PASS ✅
2. River tables never created (MessageQueue.Enabled = false)
3. validateSchema() logs: "optional tables not yet initialized"
4. System runs without River (correct behavior)
```

---

## Separation of Concerns

### Application Tables (Our Responsibility)
- Managed by our migration files (`000001_init.sql`)
- Version controlled in our repository
- We decide when to add/remove/modify
- **Must validate strictly**

### River Tables (River SDK Responsibility)
- Managed by River's internal migration system
- Version controlled in River's repository
- River decides when to add/remove/modify
- **Should not validate strictly** (we don't control them)

---

## Implementation Details

### Code Structure
```go
func validateSchema(ctx context.Context, conn *pgxpool.Conn, logger *slog.Logger) error {
    // Define table groups
    requiredTables := []string{...}  // Application tables
    optionalTables := []string{...}  // River tables
    
    // Validate required tables - HARD FAIL
    for _, table := range requiredTables {
        if !exists(table) {
            return fmt.Errorf("critical application tables missing: %v", missingRequired)
        }
    }
    
    // Validate optional tables - SOFT WARN
    for _, table := range optionalTables {
        if !exists(table) {
            logger.Info("optional tables not yet initialized", ...)
        }
    }
    
    return nil // SUCCESS even if optional tables missing
}
```

### Log Output Examples

**Fresh Database:**
```
INFO  applying migration version=000001_init
INFO  migrations completed applied=1 skipped=0 total=1
INFO  application tables validated count=7
INFO  optional tables not yet initialized tables=[river_job river_leader river_migration] note=...
INFO  schema validation completed required_tables=7 optional_tables=0
```

**Existing Database:**
```
INFO  migrations completed applied=0 skipped=1 total=1
INFO  application tables validated count=7
INFO  optional tables validated count=3
INFO  schema validation completed required_tables=7 optional_tables=3
```

**Corrupted Database:**
```
ERROR schema validation failed - application tables missing missing_tables=[instances event_outbox] hint=...
ERROR apply migrations error=critical application tables missing: [instances event_outbox]
```

---

## Benefits

✅ **Zero Breaking Changes** - No modifications needed in `main.go`  
✅ **Works in Any Environment** - Dev, staging, production, CI/CD  
✅ **Maintains Validation** - Still catches real database issues  
✅ **Clear Observability** - Logs explain exactly what's happening  
✅ **Future-Proof** - Works even if River changes table names  
✅ **Separation of Concerns** - Application and SDK tables managed independently  
✅ **Backward Compatible** - Works with existing databases  
✅ **Forward Compatible** - Works with fresh databases  

---

## Testing

### Manual Test Commands
```bash
# Test 1: Fresh database
dropdb whatsapp_api && createdb whatsapp_api
go run ./api/cmd/server/main.go
# Expected: Application starts successfully with info logs about River tables

# Test 2: Existing database with River
go run ./api/cmd/server/main.go
# Expected: Application starts with all tables validated

# Test 3: Corrupted database (simulate)
psql whatsapp_api -c "DROP TABLE instances;"
go run ./api/cmd/server/main.go
# Expected: Fatal error about missing critical tables
```

### Automated Test
```go
func TestValidateSchema(t *testing.T) {
    // Test required tables validation
    // Test optional tables soft-warning
    // Test corrupted database handling
}
```

---

## Future Considerations

### Adding New Required Tables
1. Add table creation to migration file
2. Add table name to `requiredTables` slice
3. Deploy - validation will catch missing tables

### Adding New Optional Tables
1. External system handles table creation
2. Add table name to `optionalTables` slice
3. Deploy - validation will warn if not present

### Removing Tables
1. Create new migration to drop table
2. Remove from validation slice
3. Deploy - old instances handle gracefully
