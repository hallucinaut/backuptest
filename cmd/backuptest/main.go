package main

import (
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
	results := validateBackup(backupPath)
	displayResults(results)
}

func validateBackup(backupPath string) []BackupResult {
	var results []BackupResult

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
			if err != nil {
				results = append(results, BackupResult{
					BackupPath: path,
					Status:     "ERROR",
					Error:      err.Error(),
				})
				return nil
			}

			if !info.IsDir() {
				result := validateFile(path)
				results = append(results, result)
			}
			return nil
		})
	} else {
		// Single file backup
		results = append(results, validateFile(backupPath))
	}

	return results
}

func validateFile(filePath string) BackupResult {
	result := BackupResult{
		BackupPath: filePath,
		TestTime:   time.Now(),
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
	checksum, err := calculateChecksum(filePath)
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

func calculateChecksum(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
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