package migrate

import (
	"fmt"
	"strings"

	"gorm.io/gorm"
)

// ExecuteSQLFile 执行嵌入的 SQL 文件，去除注释
func ExecuteSQLFile(tx *gorm.DB, path string) error {
	// 读取 SQL 文件内容
	sqlContent, err := sqlFiles.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read embedded SQL file: %v", err)
	}

	// 清除注释内容
	cleanedSQL := removeComments(string(sqlContent))

	// 将清除注释后的内容按分号分割成多个 SQL 语句
	sqlStatements := splitSQLStatements(cleanedSQL)

	// 遍历 SQL 语句并执行
	for _, stmt := range sqlStatements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}
		// 执行 SQL 语句
		if err := tx.Exec(stmt).Error; err != nil {
			return fmt.Errorf("failed to execute SQL statement: %v \nSQL: %s", err.Error(), stmt)
		}
	}
	return nil
}

// removeComments 去除 SQL 代码中的注释
func removeComments(sql string) string {
	var result strings.Builder
	inSingleLineComment := false
	inMultiLineComment := false
	length := len(sql)

	for i := 0; i < length; i++ {
		// 处理 -- 单行注释
		if !inMultiLineComment && !inSingleLineComment && i+1 < length && sql[i] == '-' && sql[i+1] == '-' {
			inSingleLineComment = true
			i++ // 跳过 '-'
			continue
		}

		// 结束单行注释（支持 \r\n 和 \n）
		if inSingleLineComment && (sql[i] == '\n' || sql[i] == '\r') {
			inSingleLineComment = false
		}

		// 处理 /* 多行注释 */
		if !inSingleLineComment && !inMultiLineComment && i+1 < length && sql[i] == '/' && sql[i+1] == '*' {
			inMultiLineComment = true
			i++ // 跳过 '*'
			continue
		}

		// 结束多行注释
		if inMultiLineComment && i+1 < length && sql[i] == '*' && sql[i+1] == '/' {
			inMultiLineComment = false
			i++ // 跳过 '/'
			continue
		}

		// 不是注释内容时，将字符添加到结果中
		if !inSingleLineComment && !inMultiLineComment {
			result.WriteByte(sql[i])
		}
	}

	// 去除多余的空白行
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

// splitSQLStatements 更安全地分割 SQL 语句
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
