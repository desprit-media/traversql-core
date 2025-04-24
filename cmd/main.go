package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/urfave/cli/v3"

	"github.com/desprit-media/traversql-core/internal/db"
	"github.com/desprit-media/traversql-core/internal/parser"
)

func main() {
	// Example:
	// >>> go run cmd/cli/main.go traverse --table=orders --record-id=123
	cmd := &cli.Command{
		Name:  "traverse",
		Usage: "traverse table and extract the given record and its related records",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "table",
				Usage:    "name of the table to start traversing",
				Required: true,
			},
			&cli.StringFlag{
				Name:  "schema",
				Value: "public",
				Usage: "schema of the given table",
			},
			&cli.StringSliceFlag{
				Name:    "primary-key-fields",
				Aliases: []string{"pk-fields"},
				Value:   []string{"id"},
				Usage:   "names of the fields that form the primary key of the record",
			},
			&cli.StringSliceFlag{
				Name:     "primary-key-values",
				Aliases:  []string{"pk-values"},
				Usage:    "values of the primary key of the record to start traversing",
				Required: true,
			},
			&cli.StringFlag{
				Name:  "output",
				Usage: "file to write the output to",
			},
			&cli.StringSliceFlag{
				Name:  "included-tables",
				Usage: "names of the tables to include in the traversal",
			},
			&cli.StringSliceFlag{
				Name:  "excluded-tables",
				Usage: "names of the tables to exclude from the traversal",
			},
			&cli.StringSliceFlag{
				Name:  "included-schemas",
				Usage: "names of the schemas to include in the traversal",
			},
			&cli.BoolFlag{
				Name:  "follow-parents",
				Value: true,
				Usage: "whether to follow parent relationships",
			},
			&cli.BoolFlag{
				Name:  "follow-children",
				Value: true,
				Usage: "whether to follow child relationships",
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			table := c.String("table")
			schema := c.String("schema")
			pkFields := c.StringSlice("primary-key-fields")
			pkValues := c.StringSlice("primary-key-values")
			includedTables := c.StringSlice("included-tables")
			excludedTables := c.StringSlice("excluded-tables")
			includedSchemas := c.StringSlice("included-schemas")
			followParents := c.Bool("follow-parents")
			followChildren := c.Bool("follow-children")

			columns := make([]parser.Column, len(pkFields))
			for i, pkField := range pkFields {
				columns[i] = parser.Column{Name: pkField}
			}
			values := make([]interface{}, len(pkFields))
			for i, pkValue := range pkValues {
				pkValueInt, err := strconv.Atoi(pkValue)
				if err != nil {
					values[i] = pkValue
				} else {
					values[i] = pkValueInt
				}
			}

			exists := false
			for _, s := range includedSchemas {
				if s == schema {
					exists = true
					break
				}
			}
			if !exists {
				includedSchemas = append(includedSchemas, schema)
			}

			pgConfig, err := db.GetPostgresConfig()
			if err != nil {
				return fmt.Errorf("failed to get Postgres config: %v", err)
			}
			pgPool, err := db.InitPostgresPool(ctx, pgConfig)
			if err != nil {
				return fmt.Errorf("failed to initialize Postgres pool: %v", err)
			}

			p, err := parser.NewParser(pgPool, parser.NewParserConfig(
				parser.WithSchemas(includedSchemas),
				parser.WithIncludedTables(includedTables),
				parser.WithExcludedTables(excludedTables),
				parser.WithFollowParents(followParents),
				parser.WithFollowChildren(followChildren),
			))
			if err != nil {
				return fmt.Errorf("failed to initialize parser: %v", err)
			}

			pk, err := parser.NewPrimaryKey(columns, values)
			if err != nil {
				return fmt.Errorf("failed to create primary key: %v", err)
			}
			relations, err := p.ExtractGraph(ctx, parser.Table{Name: table, Schema: schema}, pk)
			if err != nil {
				return fmt.Errorf("failed to extract records graph: %v", err)
			}
			fmt.Printf("relations:\n%s\n", relations)

			return nil
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
