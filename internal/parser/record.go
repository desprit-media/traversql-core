package parser

import (
	"context"
	"fmt"
	"reflect"
	"strings"
)

// Record represents a database record with its table and column values
type Record struct {
	Table   Table
	Columns []Column
	Values  []interface{}
}

type RecordVisit struct {
	Query string
	Args  []interface{}
}

func (r Record) String() string {
	return fmt.Sprintf("{%s %+v %+v}", r.Table.FullName(), r.Columns, r.Values)
}

// Equal compares two Records for equality
func (r Record) Equal(other Record) bool {
	if r.Table.FullName() != other.Table.FullName() {
		return false
	}

	if len(r.Columns) != len(other.Columns) || len(r.Values) != len(other.Values) {
		return false
	}

	// Build maps for comparison
	rMap := make(map[string]interface{})
	for i, col := range r.Columns {
		if i < len(r.Values) {
			rMap[col.Name] = r.Values[i]
		}
	}

	otherMap := make(map[string]interface{})
	for i, col := range other.Columns {
		if i < len(other.Values) {
			otherMap[col.Name] = other.Values[i]
		}
	}

	// Compare maps
	if len(rMap) != len(otherMap) {
		return false
	}

	for k, v := range rMap {
		otherV, ok := otherMap[k]
		if !ok || !reflect.DeepEqual(v, otherV) {
			return false
		}
	}

	return true
}

// Represents a primary key, which can be composite
type PrimaryKey struct {
	Columns []Column
	Values  []interface{}
}

// Creates a new primary key
func NewPrimaryKey(columns []Column, values []interface{}) (PrimaryKey, error) {
	if len(columns) != len(values) {
		return PrimaryKey{}, fmt.Errorf("columns and values must have the same length")
	}
	for i, col := range columns {
		if col.DataType == "" {
			columns[i].DataType = "integer"
		}
	}
	return PrimaryKey{Columns: columns, Values: values}, nil
}

// Creates a WHERE clause for a primary key
func (pk PrimaryKey) WhereClause() (string, []interface{}) {
	conditions := make([]string, len(pk.Columns))
	args := make([]interface{}, len(pk.Values))

	for i, col := range pk.Columns {
		if allowedDataTypes[col.DataType] {
			conditions[i] = fmt.Sprintf("%s = $%d", col.Name, i+1)
		} else {
			conditions[i] = fmt.Sprintf("%s::text = $%d", col.Name, i+1)
		}
		args[i] = pk.Values[i]
	}

	return strings.Join(conditions, " AND "), args
}

func (p *Parser) buildSelectColumnsQueryPart(columns []Column) string {
	selectColumns := make([]string, len(columns))
	for i, col := range columns {
		if col.DataType == "json" || col.DataType == "jsonb" ||
			col.DataType == "uuid" || col.DataType == "tsvector" {
			selectColumns[i] = fmt.Sprintf("%s::text AS %s", col.Name, col.Name)
		} else {
			selectColumns[i] = col.Name
		}
	}

	return strings.Join(selectColumns, ", ")
}

func (p *Parser) hasRecordVisit(query string, args []interface{}) bool {
	for _, visit := range p.RecordVisits {
		if visit.Query == query {
			// Compare args if they exist
			if len(visit.Args) == len(args) {
				argsMatch := true
				for i, arg := range visit.Args {
					if arg != args[i] {
						argsMatch = false
						break
					}
				}
				if argsMatch {
					return true
				}
			}
		}
	}
	return false
}

// Helper to fetch a single record by primary key
func (p *Parser) FetchRecord(ctx context.Context, table Table, pk PrimaryKey) (Record, error) {
	columns, err := p.getColumnsForTable(ctx, table)
	if err != nil {
		return Record{}, fmt.Errorf("failed to get columns for table %s: %w", table.FullName(), err)
	}

	whereClause, args := pk.WhereClause()
	selectColumns := p.buildSelectColumnsQueryPart(columns)
	query := fmt.Sprintf("SELECT %s FROM %s WHERE %s", selectColumns, table.FullName(), whereClause)
	p.logger.Printf("FIND RECORD QUERY: %s, pk: %+v", query, args)

	if !p.hasRecordVisit(query, args) {
		p.RecordVisits = append(p.RecordVisits, RecordVisit{Query: query, Args: args})
	} else {
		return Record{}, ErrRecordAlreadyVisited
	}

	row := p.pool.QueryRow(ctx, query, args...)

	// Create a slice to hold the values
	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))
	for i := range values {
		valuePtrs[i] = &values[i]
	}

	// Scan the row into the values
	if err := row.Scan(valuePtrs...); err != nil {
		return Record{}, fmt.Errorf("failed to scan record: %w", err)
	}

	return Record{
		Table:   table,
		Columns: columns,
		Values:  values,
	}, nil
}

// List of data types that can be scanned directly without casting
var allowedDataTypes = map[string]bool{
	// Core numeric types
	"integer":          true,
	"smallint":         true,
	"bigint":           true,
	"numeric":          true,
	"real":             true,
	"double precision": true,

	// Text types
	"varchar": true,
	"text":    true,
	"char":    true,

	// Boolean
	"boolean": true,

	// Date/time types
	"date":                        true,
	"time":                        true,
	"timestamp":                   true,
	"timestamp with time zone":    true,
	"timestamp without time zone": true,
	"time with time zone":         true,
	"time without time zone":      true,

	// Binary data
	"bytea": true,
}

// Helper function to find all child records that reference the parent's PK
func (p *Parser) findChildRecords(ctx context.Context, rel Relationship, parentPKValues []interface{}) ([]Record, error) {
	// Build WHERE clause for the foreign key columns
	conditions := make([]string, len(rel.SourceColumn))
	for i, col := range rel.SourceColumn {
		conditions[i] = fmt.Sprintf("%s = $%d", col.Name, i+1)
	}

	selectColumns := make([]string, len(rel.SourceTable.Columns))

	for i, col := range rel.SourceTable.Columns {
		if allowedDataTypes[col.DataType] {
			selectColumns[i] = col.Name
		} else {
			selectColumns[i] = fmt.Sprintf("%s::text AS %s", col.Name, col.Name)
		}
	}
	query := fmt.Sprintf(`SELECT %s FROM %s WHERE %s`,
		strings.Join(selectColumns, ", "),
		rel.SourceTable.FullName(),
		strings.Join(conditions, " AND "))
	p.logger.Printf("FIND CHILD RECORDS QUERY: %s, pk: %+v", query, parentPKValues)

	rows, err := p.pool.Query(ctx, query, parentPKValues...)
	if err != nil {
		return nil, fmt.Errorf("failed to query child records: %w", err)
	}
	defer rows.Close()

	var childRecords []Record

	// Get column information for the child table
	columns, err := p.getColumnsForTable(ctx, rel.SourceTable)
	if err != nil {
		return nil, fmt.Errorf("failed to get child table columns: %w", err)
	}

	for rows.Next() {
		// Create value pointers for scanning
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(values))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, fmt.Errorf("failed to scan child record: %w", err)
		}

		childRecords = append(childRecords, Record{
			Table:   rel.SourceTable,
			Columns: columns,
			Values:  values,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating child records: %w", err)
	}

	return childRecords, nil
}

func (p *Parser) tryAddRecord(records *[]Record, newRecord Record) {
	if !p.recordExists(*records, newRecord) {
		*records = append(*records, newRecord)
	}
}

// Helper to check if a record already exists in the collection
func (p *Parser) recordExists(records []Record, newRecord Record) bool {
	for _, r := range records {
		if r.Table.FullName() == newRecord.Table.FullName() {
			// Compare primary key values
			match := true
			for i, col := range r.Columns {
				if col.IsPrimary {
					if r.Values[i] != newRecord.Values[i] {
						match = false
						break
					}
				}
			}
			if match {
				return true
			}
		}
	}
	return false
}
