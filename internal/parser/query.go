package parser

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

func formatPgTime(t pgtype.Time) string {
	// Handle NULL case
	if !t.Valid {
		return "00:00:00"
	}

	// Convert microseconds to duration
	duration := time.Duration(t.Microseconds) * time.Microsecond

	// Create base time at midnight
	baseTime := time.Date(0, 1, 1, 0, 0, 0, 0, time.UTC)

	// Add duration to base time
	resultTime := baseTime.Add(duration)

	// Format as hh:mm:ss
	return resultTime.Format("15:04:05")
}

// GenerateInsertStatements generates SQL INSERT statements for the given records
func (p *Parser) GenerateInsertStatements(ctx context.Context, records []Record) (string, error) {
	var sb strings.Builder

	for _, record := range records {
		columnNames := make([]string, len(record.Columns))
		for i, col := range record.Columns {
			columnNames[i] = col.Name
		}
		columnList := strings.Join(columnNames, ", ")

		// Build value placeholders and collect values
		var placeholders []string
		var values []string

		for i, val := range record.Values {
			if val == nil {
				values = append(values, "NULL")
			} else {
				switch v := val.(type) {
				case string:
					values = append(values, fmt.Sprintf("'%s'", strings.ReplaceAll(v, "'", "''")))
				case []byte:
					values = append(values, fmt.Sprintf("'%s'", strings.ReplaceAll(string(v), "'", "''")))
				case int32:
					values = append(values, fmt.Sprintf("%d", v))
				case int16:
					values = append(values, fmt.Sprintf("%d", v))
				case pgtype.Time:
					values = append(values, fmt.Sprintf("'%s'", formatPgTime(v)))
				case pgtype.Numeric:
					n, _ := v.Float64Value()
					values = append(values, fmt.Sprintf("%g", n.Float64))
				case time.Time:
					// Format time as ISO 8601 string
					values = append(values, fmt.Sprintf("'%s'", v.Format("2006-01-02T15:04:05.999999Z07:00")))
				case bool:
					if v {
						values = append(values, "true")
					} else {
						values = append(values, "false")
					}
				case float32:
					values = append(values, fmt.Sprintf("%g", v))
				case float64:
					values = append(values, fmt.Sprintf("%g", v))
				case int:
					values = append(values, fmt.Sprintf("%d", v))
				case int64:
					values = append(values, fmt.Sprintf("%d", v))
				case map[string]interface{}:
					// Handle JSON data (OID 3614)
					jsonBytes, err := json.Marshal(v)
					if err != nil {
						p.logger.Printf("Error marshaling JSON: %v\n", err)
						values = append(values, fmt.Sprintf("'%v'", v))
					} else {
						values = append(values, fmt.Sprintf("'%s'", strings.ReplaceAll(string(jsonBytes), "'", "''")))
					}
				case []interface{}:
					// Handle JSON array
					jsonBytes, err := json.Marshal(v)
					if err != nil {
						p.logger.Printf("Error marshaling JSON array: %v\n", err)
						values = append(values, fmt.Sprintf("'%v'", v))
					} else {
						values = append(values, fmt.Sprintf("'%s'", strings.ReplaceAll(string(jsonBytes), "'", "''")))
					}
				default:
					p.logger.Printf("Unknown type: %T\n", v)
					values = append(values, fmt.Sprintf("%v", v))
				}
			}
			placeholders = append(placeholders, fmt.Sprintf("$%d", i+1))
		}

		valueList := strings.Join(values, ", ")

		// Build the INSERT statement
		stmt := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s);\n",
			record.Table.FullName(), columnList, valueList)

		sb.WriteString(stmt)
	}

	return sb.String(), nil
}
