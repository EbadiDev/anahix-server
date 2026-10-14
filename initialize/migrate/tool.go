package migrate

import (
	"fmt"
	"strings"

	"gorm.io/gorm"
)

// ExecuteSQLFile executes an embedded SQL migration file, stripping comments and handling statements
func ExecuteSQLFile(tx *gorm.DB, path string) error {
	sqlContent, err := sqlFiles.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read embedded SQL file %s: %w", path, err)
	}

	cleanedSQL := removeComments(string(sqlContent))
	sqlStatements := splitSQLStatements(cleanedSQL)

	for _, stmt := range sqlStatements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}
		if err := tx.Exec(stmt).Error; err != nil {
			return fmt.Errorf("failed to execute SQL statement: %w\nSQL:\n%s", err, stmt)
		}
	}
	return nil
}

// removeComments strips single-line (--) and multi-line (/* ... */) comments from SQL
func removeComments(sql string) string {
	var result strings.Builder
	inSingleLineComment := false
	inMultiLineComment := false
	length := len(sql)

	for i := 0; i < length; i++ {
		// Single-line comment --
		if !inMultiLineComment && !inSingleLineComment && i+1 < length && sql[i] == '-' && sql[i+1] == '-' {
			inSingleLineComment = true
			i++
			continue
		}

		// End of single-line comment
		if inSingleLineComment && (sql[i] == '\n' || sql[i] == '\r') {
			inSingleLineComment = false
		}

		// Multi-line comment /*
		if !inSingleLineComment && !inMultiLineComment && i+1 < length && sql[i] == '/' && sql[i+1] == '*' {
			inMultiLineComment = true
			i++
			continue
		}

		// End of multi-line comment */
		if inMultiLineComment && i+1 < length && sql[i] == '*' && sql[i+1] == '/' {
			inMultiLineComment = false
			i++
			continue
		}

		if !inSingleLineComment && !inMultiLineComment {
			result.WriteByte(sql[i])
		}
	}

	lines := strings.Split(result.String(), "\n")
	var cleanedLines []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			cleanedLines = append(cleanedLines, trimmed)
		}
	}
	return strings.Join(cleanedLines, "\n")
}

// splitSQLStatements safely splits SQL by semicolon
func splitSQLStatements(sql string) []string {
	statements := strings.Split(sql, ";")
	var results []string
	for _, stmt := range statements {
		trimmed := strings.TrimSpace(stmt)
		if trimmed != "" {
			results = append(results, trimmed)
		}
	}
	return results
}
