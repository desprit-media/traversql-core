package parser

import "fmt"

var (
	ErrNoTablesFound        = fmt.Errorf("no tables found")
	ErrNoRelationshipsFound = fmt.Errorf("no relationships found")
	ErrNoPrimaryKeyFound    = fmt.Errorf("no primary key found")
	ErrRecordAlreadyVisited = fmt.Errorf("record already visited")
)
