package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/urfave/cli/v3"

	"github.com/desprit-media/traversql-core/internal/db"
	"github.com/desprit-media/traversql-core/internal/parser"
)

// createPK constructs a parser.PrimaryKey from the provided primary key fields and values.
// It attempts to convert values to integers if possible.
// pkFields: Slice of primary key field names.
// pkValues: Slice of primary key values (as strings).
func createPK(pkFields []string, pkValues []string) (parser.PrimaryKey, error) {
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

	pk, err := parser.NewPrimaryKey(columns, values)
	if err != nil {
		return parser.PrimaryKey{}, fmt.Errorf("failed to create primary key: %v", err)
	}

	return pk, nil
}

// createPgPool initializes a new PostgreSQL connection pool using configuration from environment variables.
// ctx: The context for the pool initialization.
func createPgPool(ctx context.Context) (*pgxpool.Pool, error) {
	pgConfig, err := db.NewPostgresConfigFromEnvs()
	if err != nil {
		return nil, fmt.Errorf("failed to get Postgres config: %v", err)
	}
	pgPool, err := db.InitPostgresPool(ctx, pgConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Postgres pool: %v", err)
	}

	return pgPool, err
}

// createIncludedSchemas ensures that the schema of the starting table is included in the list of schemas to traverse.
// includedSchemas: The initial slice of included schemas from command line flags.
// schema: The schema of the starting table.
// Returns the updated slice of included schemas.
func createIncludedSchemas(includedSchemas []string, schema string) []string {
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

	return includedSchemas
}

// writeGraph writes the provided graph string to either a specified output file or standard output.
// outputFileName: The name of the file to write to. If empty, output goes to standard output.
// graph: The string representation of the graph to write.
func writeGraph(outputFileName string, graph string) error {
	var outputWriter io.Writer = os.Stdout // Default to standard output

	if outputFileName != "" {
		outputFile, err := os.Create(outputFileName)
		if err != nil {
			return fmt.Errorf("failed to create output file %s: %v", outputFileName, err)
		}
		defer outputFile.Close()

		outputWriter = outputFile
	}

	fmt.Fprintln(outputWriter, graph)

	return nil
}

func main() {
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
			pgPool, err := createPgPool(ctx)
			if err != nil {
				return fmt.Errorf("failed to create Postgres pool: %v", err)
			}

			includedSchemas := createIncludedSchemas(c.StringSlice("included-schemas"), c.String("schema"))

			// Construct primary key for the value we use to start traversing
			pk, err := createPK(c.StringSlice("primary-key-fields"), c.StringSlice("primary-key-values"))
			if err != nil {
				return fmt.Errorf("failed to create primary key: %v", err)
			}

			p, err := parser.NewParser(pgPool, parser.NewParserConfig(
				parser.WithSchemas(includedSchemas),
				parser.WithIncludedTables(c.StringSlice("included-tables")),
				parser.WithExcludedTables(c.StringSlice("excluded-tables")),
				parser.WithFollowParents(c.Bool("follow-parents")),
				parser.WithFollowChildren(c.Bool("follow-children")),
			))
			if err != nil {
				return fmt.Errorf("failed to initialize parser: %v", err)
			}

			graph, err := p.ExtractGraph(ctx, parser.Table{Name: c.String("table"), Schema: c.String("schema")}, pk)
			if err != nil {
				return fmt.Errorf("failed to extract records graph: %v", err)
			}

			if err := writeGraph(c.String("output"), graph); err != nil {
				return fmt.Errorf("failed to write graph: %v", err)
			}

			return nil
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
