package parser

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Main service for extraction
type Parser struct {
	pool   *pgxpool.Pool
	config *parserConfig
	logger *log.Logger

	TablesWithPrimaryKey    []Table
	TablesWithoutPrimaryKey []Table
	Relationships           []Relationship
	TableToPKColumnsMap     map[string][]Column
	RelationshipVisits      []RelationshipVisit
	RecordVisits            []RecordVisit
}

// Initialize a new parser
func NewParser(pool *pgxpool.Pool, config *parserConfig) (*Parser, error) {
	p := &Parser{
		pool:   pool,
		config: config,
		logger: log.Default(),

		TablesWithPrimaryKey:    make([]Table, 0),
		TablesWithoutPrimaryKey: make([]Table, 0),
		Relationships:           make([]Relationship, 0),
		TableToPKColumnsMap:     make(map[string][]Column),
		RelationshipVisits:      make([]RelationshipVisit, 0),
		RecordVisits:            make([]RecordVisit, 0),
	}
	ctx := context.Background()

	tablesWithPK, tablesWithoutPK, err := p.discoverTables(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to extract tables: %w", err)
	}
	if len(tablesWithPK) == 0 {
		return nil, ErrNoTablesFound
	}
	p.TablesWithPrimaryKey = tablesWithPK
	p.TablesWithoutPrimaryKey = tablesWithoutPK

	relationships, err := p.discoverRelationships(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to extract relationships: %w", err)
	}
	if len(relationships) == 0 {
		return nil, ErrNoRelationshipsFound
	}
	p.Relationships = relationships

	for _, table := range p.TablesWithPrimaryKey {
		pk, err := p.discoverTablePKColumn(table)
		if err != nil {
			return nil, fmt.Errorf("failed to extract primary keys for table %s: %w", table.FullName(), err)
		}
		p.TableToPKColumnsMap[table.FullName()] = pk
	}

	return p, nil
}

func (p *Parser) BuildGraph(ctx context.Context, table Table, pk PrimaryKey) ([]Record, error) {
	var records []Record

	record, err := p.FetchRecord(ctx, table, pk)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch entry record: %w", err)
	}

	if p.config.FollowParents {
		if err := p.TraverseParents(ctx, record, &records); err != nil {
			return nil, fmt.Errorf("failed to traverse parents: %w", err)
		}
	}

	records = append(records, record)

	if p.config.FollowChildren {
		if err := p.TraverseChildren(ctx, record, &records); err != nil {
			return nil, fmt.Errorf("failed to traverse children: %w", err)
		}
	}

	return records, nil
}

func (p *Parser) ExtractGraph(ctx context.Context, table Table, pk PrimaryKey) (string, error) {
	defer p.Reset()

	records, err := p.BuildGraph(ctx, table, pk)
	if err != nil {
		return "", fmt.Errorf("failed to build graph: %w", err)
	}

	sql, err := p.GenerateInsertStatements(ctx, records)
	if err != nil {
		return "", fmt.Errorf("failed to generate insert statements: %w", err)
	}

	return sql, nil
}

func (p *Parser) Reset() {
	p.RelationshipVisits = make([]RelationshipVisit, 0)
	p.RecordVisits = make([]RecordVisit, 0)
}

// TraverseParents gets all relationships where table of the given record is the source (child)
func (p *Parser) TraverseParents(ctx context.Context, record Record, records *[]Record) error {
	//
	// For example in a schema with `users` -> `orders` tables
	// when we use `orders` as an entry record, we want to find relationships
	// which have `orders` as a child dependency, so we look for `users` table.
	//
	for _, rel := range p.Relationships {
		if rel.SourceTable.FullName() == record.Table.FullName() {
			// Skip if we've already visited this relationship
			// if p.hasRelationshipVisit(rel.SourceTable, rel.TargetTable) {
			// 	continue
			// }
			// Mark this relationship as visited
			p.addRelationshipVisit(rel.SourceTable, rel.TargetTable)

			// Find the value of the foreign key column in our current record
			var fkValues []interface{}
			for _, relCol := range rel.SourceColumn {
				for i, col := range record.Columns {
					if col.Name == relCol.Name {
						if record.Values[i] == nil {
							continue
						}
						fkValues = append(fkValues, record.Values[i])
						break
					}
				}
			}
			if len(fkValues) == 0 {
				continue // No foreign key value found
			}

			// Create a primary key for the parent table
			pk, err := NewPrimaryKey(rel.TargetColumn, fkValues)
			if err != nil {
				return fmt.Errorf("failed to create primary key for parent table: %w", err)
			}

			// Fetch the parent record
			parentRecord, err := p.FetchRecord(ctx, rel.TargetTable, pk)
			if err != nil {
				if errors.Is(err, ErrRecordAlreadyVisited) {
					p.logger.Printf("skipping already visited record in table %s, pk %v", rel.TargetTable.Name, pk.Values)
					continue
				}
				return fmt.Errorf("failed to fetch parent record: %w", err)
			}

			// Add to our records and keep traversing if this record hasn't been checked yet
			if !p.recordExists(*records, parentRecord) {
				// Recursively traverse parents of this parent
				if err := p.TraverseParents(ctx, parentRecord, records); err != nil {
					return err
				}
				*records = append(*records, parentRecord)
			}
		}
	}
	return nil
}

// TraverseChildren gets all relationships where table of the given record is the target (parent)
func (p *Parser) TraverseChildren(ctx context.Context, record Record, records *[]Record) error {
	// Get the primary key values of the current record
	pkValues := make([]interface{}, 0)
	for i, col := range record.Columns {
		if col.IsPrimary {
			pkValues = append(pkValues, record.Values[i])
			continue
		}
	}
	if len(pkValues) == 0 {
		return nil // No primary key, can't find children
	}

	// Get all relationships where this table is the target (parent)
	for _, rel := range p.Relationships {
		if rel.TargetTable.FullName() == record.Table.FullName() {
			// Skip if we've already visited this relationship
			// if p.hasRelationshipVisit(rel.TargetTable, rel.SourceTable) {
			// 	continue
			// }

			// Mark this relationship as visited
			p.addRelationshipVisit(rel.TargetTable, rel.SourceTable)

			// Find all child records that reference this record's primary key
			childRecords, err := p.findChildRecords(ctx, rel, pkValues)
			if err != nil {
				return fmt.Errorf("failed to find child records: %w", err)
			}

			for _, childRecord := range childRecords {
				// Add to our records if not already present
				if !p.recordExists(*records, childRecord) {
					if p.config.FollowParents {
						if err := p.TraverseParents(ctx, childRecord, records); err != nil {
							return fmt.Errorf("failed to traverse parents of child %s: %w", childRecord.Table.FullName(), err)
						}
					}

					*records = append(*records, childRecord)

					// Recursively traverse children of this child
					if p.config.FollowChildren {
						if err := p.TraverseChildren(ctx, childRecord, records); err != nil {
							return err
						}
					}
				}
			}
		}
	}
	return nil
}
