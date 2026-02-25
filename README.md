# backuptest

Backup Integrity Validator - validates backup file integrity through checksum verification.

## Purpose

Verify backup files and directories are intact and readable. Calculates MD5 checksums for integrity verification.

## Installation

```bash
go build -o backuptest ./cmd/backuptest
```

## Usage

```bash
backuptest <backup_path>
```

### Examples

```bash
# Validate single file
backuptest /backup/daily/database.sql

# Validate entire backup directory
backuptest /backup/daily

# Validate compressed archive
backuptest /backup/weekly/backup.tar.gz
```

## Output

```
=== BACKUP INTEGRITY TEST RESULTS ===

[OK] /backup/daily/database.sql
    Size: 1.2 GB | Checksum: a3f5b8c2d9e1f4a6b7c8d9e0f1a2b3c4

=== SUMMARY ===
  Valid: 1
  Warnings: 0
  Errors: 0

Backup integrity verified successfully!
```

## Status Codes

- OK: File is valid and readable
- WARNING: File exists but is empty (0 bytes)
- ERROR: File cannot be accessed or read

## Dependencies

- Go 1.21+
- github.com/fatih/color

## Build and Run

```bash
# Build
go build -o backuptest ./cmd/backuptest

# Run
go run ./cmd/backuptest /path/to/backup
```

## License

MIT