package main

import (
	"os/signal"
	"syscall"
	"context"
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/fatih/color"
)

type BackupResult struct {
	BackupPath string
	Size       int64
	Checksum   string
	Status     string
	Error      string
	TestTime   time.Time
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		cancel()
	}()

	if len(os.Args) < 2 {
		fmt.Println(color.CyanString("backuptest - Backup Integrity Validator"))
		fmt.Println()
		fmt.Println("Usage: backuptest <backup_path>")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  backuptest /backup/daily")
		fmt.Println("  backuptest /backup/daily/database.sql")
		os.Exit(1)
	}

	backupPath := os.Args[1]
	results := validateBackup(ctx, backupPath)
	displayResults(results)
}

func validateBackup(ctx context.Context, backupPath string) []BackupResult {
	var results []BackupResult

	select {
	case <-ctx.Done():
		results = append(results, BackupResult{
			BackupPath: backupPath,
			Status:     "ERROR",
			Error:      "context cancelled",
		})
		return results
	default:
	}

	info, err := os.Stat(backupPath)
	if err != nil {
		results = append(results, BackupResult{
			BackupPath: backupPath,
			Status:     "ERROR",
			Error:      err.Error(),
		})
		return results
	}

	if info.IsDir() {
		// Directory backup - validate all files
		filepath.Walk(backupPath, func(path string, info os.FileInfo, err error) error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			if err != nil {
				results = append(results, BackupResult{
					BackupPath: path,
					Status:     "ERROR",
					Error:      err.Error(),
				})
				return nil
			}

			if !info.IsDir() {
				result := validateFile(ctx, path)
				results = append(results, result)
			}
			return nil
		})
	} else {
		// Single file backup
		results = append(results, validateFile(ctx, backupPath))
	}

	return results
}

func validateFile(ctx context.Context, filePath string) BackupResult {
	result := BackupResult{
		BackupPath: filePath,
		TestTime:   time.Now(),
	}

	select {
	case <-ctx.Done():
		result.Status = "ERROR"
		result.Error = "context cancelled"
		return result
	default:
	}

	// Check file exists and is readable
	file, err := os.Open(filePath)
	if err != nil {
		result.Status = "ERROR"
		result.Error = err.Error()
		return result
	}
	defer file.Close()

	// Get file size
	info, err := file.Stat()
	if err != nil {
		result.Status = "ERROR"
		result.Error = err.Error()
		return result
	}
	result.Size = info.Size()

	// Calculate checksum
	checksum, err := calculateChecksum(ctx, filePath)
	if err != nil {
		result.Status = "ERROR"
		result.Error = err.Error()
		return result
	}
	result.Checksum = checksum

	// Verify file integrity
	if result.Size == 0 {
		result.Status = "WARNING"
		result.Error = "Empty file"
	} else {
		result.Status = "OK"
	}

	return result
}

func calculateChecksum(ctx context.Context, filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

func displayResults(results []BackupResult) {
	fmt.Println(color.CyanString("\n=== BACKUP INTEGRITY TEST RESULTS ===\n"))

	var ok, warning, errorCount int

	for _, r := range results {
		statusColor := color.GreenString
		if r.Status == "WARNING" {
			statusColor = color.YellowString
			warning++
		} else if r.Status == "ERROR" {
			statusColor = color.RedString
			errorCount++
		} else {
			ok++
		}

		fmt.Printf("[%s] %s\n",
			statusColor(r.Status),
			r.BackupPath,
		)

		fmt.Printf("    Size: %s | Checksum: %s\n",
			formatSize(r.Size),
			color.HiWhiteString(r.Checksum),
		)

		if r.Error != "" {
			fmt.Printf("    %s: %s\n", color.RedString("Error"), r.Error)
		}
		fmt.Println()
	}

	fmt.Println(color.CyanString("\n=== SUMMARY ==="))
	fmt.Printf("  Valid: %d\n", ok)
	fmt.Printf("  Warnings: %d\n", warning)
	fmt.Printf("  Errors: %d\n", errorCount)

	if errorCount == 0 && warning == 0 {
		fmt.Println(color.GreenString("\n✓ Backup integrity verified successfully!"))
	}
}

func formatSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}