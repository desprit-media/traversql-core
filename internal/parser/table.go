package parser

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Represents a database column
type Column struct {
	Name      string
	DataType  string
	IsPrimary bool
}

// Represents a database table
type Table struct {
	Name    string
	Schema  string
	Columns []Column
}

func (t Table) String() string {
	return fmt.Sprintf("{%s %s %+v}", t.Name, t.Schema, t.Columns)
}

// String representation of a table with schema
func (t Table) FullName() string {
	return fmt.Sprintf("%s.%s", t.Schema, t.Name)
}

// Get columns for a table
func (t Table) getColumns(ctx context.Context, pool *pgxpool.Pool) ([]Column, error) {
	query := `
        SELECT
            column_name,
            data_type,
            (
                SELECT
                    count(*) > 0
                FROM
                    information_schema.table_constraints tc
                    JOIN information_schema.key_column_usage kcu
                        ON tc.constraint_name = kcu.constraint_name
                        AND tc.table_schema = kcu.table_schema
                        AND tc.table_name = kcu.table_name
                WHERE
                    tc.constraint_type = 'PRIMARY KEY'
                    AND tc.table_schema = $1
                    AND tc.table_name = $2
                    AND kcu.column_name = c.column_name
            ) as is_primary
        FROM
            information_schema.columns c
        WHERE
            table_schema = $1
            AND table_name = $2
        ORDER BY
            ordinal_position
    `

	rows, err := pool.Query(ctx, query, t.Schema, t.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to query columns: %w", err)
	}
	defer rows.Close()

	var columns []Column
	for rows.Next() {
		var col Column
		if err := rows.Scan(&col.Name, &col.DataType, &col.IsPrimary); err != nil {
			return nil, fmt.Errorf("failed to scan column row: %w", err)
		}
		columns = append(columns, col)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating column rows: %w", err)
	}

	return columns, nil
}

func (p *Parser) discoverTablePKColumn(table Table) ([]Column, error) {
	var pkColumns []Column
	for _, col := range table.Columns {
		if col.IsPrimary {
			pkColumns = append(pkColumns, col)
		}
	}

	if len(pkColumns) == 0 {
		return nil, fmt.Errorf("no primary key found for table %s", table.FullName())
	}

	return pkColumns, nil
}

func (p *Parser) discoverTables(ctx context.Context) ([]Table, error) {
	query := `
        SELECT
            table_schema,
            table_name
        FROM
            information_schema.tables
        WHERE
            table_schema = ANY($1)
            AND table_type = 'BASE TABLE'
    `

	rows, err := p.pool.Query(ctx, query, p.config.Schemas)
	if err != nil {
		return nil, fmt.Errorf("failed to query tables: %w", err)
	}
	defer rows.Close()

	var tables []Table
	for rows.Next() {
		var schema, name string
		if err := rows.Scan(&schema, &name); err != nil {
			return nil, fmt.Errorf("failed to scan table row: %w", err)
		}

		// Skip excluded tables
		if contains(p.config.ExcludedTables, name) {
			continue
		}
		// Skip tables not in included list (if specified)
		if len(p.config.IncludedTables) > 0 && !contains(p.config.IncludedTables, name) {
			continue
		}

		table := Table{
			Schema: schema,
			Name:   name,
		}

		columns, err := table.getColumns(ctx, p.pool)
		if err != nil {
			return nil, fmt.Errorf("failed to get columns for table %s: %w", table.FullName(), err)
		}
		table.Columns = columns

		tables = append(tables, table)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating table rows: %w", err)
	}

	return tables, nil
}

func (p *Parser) getTable(schema, name string) (Table, error) {
	for _, table := range p.Tables {
		if table.Schema == schema && table.Name == name {
			return table, nil
		}
	}
	return Table{}, fmt.Errorf("table %s.%s not found", schema, name)
}

func (p *Parser) getColumnsForTable(ctx context.Context, table Table) ([]Column, error) {
	for _, t := range p.Tables {
		if t.FullName() == table.FullName() {
			return t.Columns, nil
		}
	}
	return table.getColumns(ctx, p.pool)
}
