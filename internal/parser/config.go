package parser

// Configuration for the extraction
type parserConfig struct {
	// Schemas to extract from
	Schemas []string
	// Tables to exclude from extraction
	ExcludedTables []string
	// Tables to include (if empty, include all non-excluded)
	IncludedTables []string
	// Whether to follow parent relationships (foreign keys pointing to other tables)
	FollowParents bool
	// Whether to follow child relationships (foreign keys from other tables pointing to this one)
	FollowChildren bool
}

func NewParserConfig(opts ...ConfigOpt) *parserConfig {
	c := &parserConfig{
		Schemas:        []string{},
		ExcludedTables: []string{},
		IncludedTables: []string{},
		FollowParents:  true,
		FollowChildren: true,
	}
	for _, opt := range opts {
		opt(c)
	}
	if len(c.Schemas) == 0 {
		c.Schemas = []string{"public"}
	}
	return c
}

type ConfigOpt func(*parserConfig)

func WithSchemas(schemas []string) ConfigOpt {
	return func(c *parserConfig) {
		c.Schemas = append(c.Schemas, schemas...)
	}
}

func WithIncludedTables(tables []string) ConfigOpt {
	return func(c *parserConfig) {
		c.IncludedTables = append(c.IncludedTables, tables...)
	}
}

func WithExcludedTables(tables []string) ConfigOpt {
	return func(c *parserConfig) {
		c.ExcludedTables = append(c.ExcludedTables, tables...)
	}
}

func WithFollowParents(follow bool) ConfigOpt {
	return func(c *parserConfig) {
		c.FollowParents = follow
	}
}

func WithFollowChildren(follow bool) ConfigOpt {
	return func(c *parserConfig) {
		c.FollowChildren = follow
	}
}
