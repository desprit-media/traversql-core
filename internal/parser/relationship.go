package parser

import (
	"context"
	"fmt"
)

type RelationType string

const (
	// One-to-Many relationship
	OneToMany RelationType = "one-to-many"
	// Many-to-One relationship
	ManyToOne RelationType = "many-to-one"
	// Many-to-Many relationship
	ManyToMany RelationType = "many-to-many"
	// One-to-One relationship
	OneToOne RelationType = "one-to-one"
	// Self-Referencing relationship
	SelfReferencing RelationType = "self-referencing"
)

// Represents a foreign key relationship
type Relationship struct {
	SourceTable  Table
	SourceColumn []Column
	TargetTable  Table
	TargetColumn []Column
	RelationType RelationType
}

func (r Relationship) String() string {
	return fmt.Sprintf("%s | %s.%+v -> %s.%+v", r.RelationType, r.SourceTable.FullName(), r.SourceColumn, r.TargetTable.FullName(), r.TargetColumn)
}

// Track of visited tables and also direction of that visit
type RelationshipVisit struct {
	TableFrom Table
	TableTo   Table
}

func (p *Parser) discoverRelationships(ctx context.Context) ([]Relationship, error) {
	query := `
        SELECT
            tc.table_schema as source_schema,
            tc.table_name as source_table,
            kcu.column_name as source_column,
            ccu.table_schema as target_schema,
            ccu.table_name as target_table,
            ccu.column_name as target_column,
            tc.constraint_name,
            kcu.ordinal_position,
            (SELECT COUNT(*) FROM information_schema.table_constraints WHERE constraint_name = tc.constraint_name) as constraint_count,
            (SELECT COUNT(*) FROM information_schema.key_column_usage WHERE constraint_name = tc.constraint_name) as columns_in_constraint,
            (SELECT COUNT(*)
            FROM information_schema.table_constraints tc2
            JOIN information_schema.key_column_usage kcu2 ON tc2.constraint_name = kcu2.constraint_name
            WHERE tc2.constraint_type = 'PRIMARY KEY'
            AND kcu2.table_schema = ccu.table_schema
            AND kcu2.table_name = ccu.table_name
            AND kcu2.column_name = ccu.column_name) as is_target_pk,
            (SELECT COUNT(*)
            FROM information_schema.table_constraints tc3
            JOIN information_schema.key_column_usage kcu3 ON tc3.constraint_name = kcu3.constraint_name
            WHERE tc3.constraint_type = 'UNIQUE'
            AND kcu3.table_schema = ccu.table_schema
            AND kcu3.table_name = ccu.table_name
            AND kcu3.column_name = ccu.column_name) as is_target_unique
        FROM
            information_schema.table_constraints tc
            JOIN information_schema.key_column_usage kcu
                ON tc.constraint_name = kcu.constraint_name
                AND tc.table_schema = kcu.table_schema
                AND tc.table_name = kcu.table_name
            JOIN information_schema.constraint_column_usage ccu
                ON ccu.constraint_name = tc.constraint_name
                AND ccu.table_schema = tc.table_schema
        WHERE
            tc.constraint_type = 'FOREIGN KEY'
            AND tc.table_schema = ANY($1)
    `

	rows, err := p.pool.Query(ctx, query, p.config.Schemas)
	if err != nil {
		return nil, fmt.Errorf("failed to query relationships: %w", err)
	}
	defer rows.Close()

	var relationships []Relationship
	for rows.Next() {
		var sourceSchema, sourceTable, sourceColumn, targetSchema, targetTable, targetColumn, constraintName string
		var ordinalPosition, constraintCount, columnsInConstraint, isTargetPK, isTargetUnique int
		if err := rows.Scan(
			&sourceSchema, &sourceTable, &sourceColumn,
			&targetSchema, &targetTable, &targetColumn,
			&constraintName, &ordinalPosition, &constraintCount, &columnsInConstraint,
			&isTargetPK, &isTargetUnique,
		); err != nil {
			return nil, fmt.Errorf("failed to scan relationship row: %w", err)
		}

		// Skip if source or target table is excluded
		if contains(p.config.ExcludedTables, sourceTable) || contains(p.config.ExcludedTables, targetTable) {
			continue
		}
		// Skip if included tables are specified and either source or target is not included
		if len(p.config.IncludedTables) > 0 &&
			(!contains(p.config.IncludedTables, sourceTable) || !contains(p.config.IncludedTables, targetTable)) {
			continue
		}

		// Get source and target tables
		sourceTableObj, err := p.getTable(sourceSchema, sourceTable)
		if err != nil {
			return nil, err
		}
		targetTableObj, err := p.getTable(targetSchema, targetTable)
		if err != nil {
			return nil, err
		}

		// Find source and target columns
		var sourceCol, targetCol Column
		for _, col := range sourceTableObj.Columns {
			if col.Name == sourceColumn {
				sourceCol = col
				break
			}
		}
		for _, col := range targetTableObj.Columns {
			if col.Name == targetColumn {
				targetCol = col
				break
			}
		}

		var relType RelationType

		// Check if this is a self-referencing relationship
		if sourceTable == targetTable {
			relType = SelfReferencing
		} else {
			// Check if target column is a primary key or has a unique constraint
			isTargetKeyUnique := isTargetPK > 0 || isTargetUnique > 0

			// Get information about source column uniqueness
			var isSourceKeyUnique bool
			err := p.pool.QueryRow(ctx, `
               SELECT COUNT(*) > 0
               FROM information_schema.table_constraints tc
               JOIN information_schema.key_column_usage kcu ON tc.constraint_name = kcu.constraint_name
               WHERE (tc.constraint_type = 'PRIMARY KEY' OR tc.constraint_type = 'UNIQUE')
               AND kcu.table_schema = $1 AND kcu.table_name = $2 AND kcu.column_name = $3
           `, sourceSchema, sourceTable, sourceColumn).Scan(&isSourceKeyUnique)

			if err != nil {
				return nil, fmt.Errorf("failed to check source column uniqueness: %w", err)
			}

			// Determine relationship type based on column uniqueness
			if isSourceKeyUnique && isTargetKeyUnique {
				relType = OneToOne
			} else if isSourceKeyUnique && !isTargetKeyUnique {
				relType = OneToMany
			} else if !isSourceKeyUnique && isTargetKeyUnique {
				relType = ManyToOne
			} else {
				relType = ManyToMany
			}
		}

		rel := Relationship{
			SourceTable:  sourceTableObj,
			SourceColumn: []Column{sourceCol},
			TargetTable:  targetTableObj,
			TargetColumn: []Column{targetCol},
			RelationType: relType,
		}

		relationships = append(relationships, rel)

		// // Also add an inverse relationship for the target table
		// var inverseRelType RelationType
		// switch relType {
		// case OneToMany:
		// 	inverseRelType = ManyToOne
		// case ManyToOne:
		// 	inverseRelType = OneToMany
		// case OneToOne:
		// 	inverseRelType = OneToOne
		// case ManyToMany:
		// 	inverseRelType = ManyToMany
		// case SelfReferencing:
		// 	inverseRelType = SelfReferencing
		// }
		// inverseRel := Relationship{
		// 	SourceTable:  targetTableObj,
		// 	SourceColumn: []Column{targetCol},
		// 	TargetTable:  sourceTableObj,
		// 	TargetColumn: []Column{sourceCol},
		// 	RelationType: inverseRelType,
		// }

		// relationships = append(relationships, inverseRel)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating relationship rows: %w", err)
	}

	return relationships, nil
}

func (s *Parser) addRelationshipVisit(from, to Table) {
	fmt.Printf("ADDING VISIT FROM %s TO %s\n", from.FullName(), to.FullName())
	s.RelationshipVisits = append(s.RelationshipVisits, RelationshipVisit{TableFrom: from, TableTo: to})
}

func (s *Parser) hasRelationshipVisit(from, to Table) bool {
	fmt.Printf("CHECKING VISIT FROM %s TO %s\n", from.FullName(), to.FullName())
	for _, visit := range s.RelationshipVisits {
		if visit.TableFrom.Name == from.Name && visit.TableTo.Name == to.Name {
			return true
		}
	}
	return false
}
