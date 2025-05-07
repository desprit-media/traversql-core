package tests

import (
	"context"
	"math/big"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"

	"github.com/desprit-media/traversql-core/internal/parser"
)

func TestParser(t *testing.T) {
	ctx := context.Background()

	t.Run("should discover tables", func(t *testing.T) {
		cases := []struct {
			name     string
			mocks    []string
			schemas  []string
			expected []string
		}{
			{
				name:    "one-to-one",
				mocks:   []string{"001_one_to_one/001_tables.sql", "001_one_to_one/002_records.sql"},
				schemas: []string{"public"},
				expected: []string{
					"{users public [{Name:id DataType:integer IsPrimary:true} {Name:name DataType:character varying IsPrimary:false}]}",
					"{orders public [{Name:id DataType:integer IsPrimary:true} {Name:user_id DataType:integer IsPrimary:false} {Name:amount DataType:numeric IsPrimary:false}]}",
					"{payments public [{Name:id DataType:integer IsPrimary:true} {Name:order_id DataType:integer IsPrimary:false} {Name:amount DataType:numeric IsPrimary:false}]}",
				},
			},
			{
				name:    "many-to-many",
				mocks:   []string{"002_many_to_many/001_tables.sql", "002_many_to_many/002_records.sql"},
				schemas: []string{"public"},
				expected: []string{
					"{users public [{Name:id DataType:integer IsPrimary:true} {Name:name DataType:character varying IsPrimary:false}]}",
					"{user_orders public [{Name:user_id DataType:integer IsPrimary:true} {Name:order_id DataType:integer IsPrimary:true}]}",
					"{orders public [{Name:id DataType:integer IsPrimary:true} {Name:amount DataType:numeric IsPrimary:false}]}",
					"{order_payments public [{Name:order_id DataType:integer IsPrimary:true} {Name:payment_id DataType:integer IsPrimary:true}]}",
					"{payments public [{Name:payment_id DataType:integer IsPrimary:true} {Name:amount DataType:numeric IsPrimary:false}]}",
				},
			},
			{
				name:    "self-referencing",
				mocks:   []string{"005_self_referencing/001_tables.sql", "005_self_referencing/002_records.sql"},
				schemas: []string{"public"},
				expected: []string{
					"{persons public [{Name:person_id DataType:integer IsPrimary:true} {Name:first_name DataType:character varying IsPrimary:false} {Name:gender_id DataType:integer IsPrimary:false} {Name:parent_id DataType:integer IsPrimary:false}]}",
					"{genders public [{Name:gender_id DataType:integer IsPrimary:true} {Name:gender_name DataType:character varying IsPrimary:false}]}",
				},
			},
		}

		for _, c := range cases {
			_, pgPool := NewPostgresContainer(ctx, t, c.mocks...)
			p, err := parser.NewParser(pgPool, parser.NewParserConfig(parser.WithSchemas(c.schemas)))
			if assert.NoError(t, err, "failed to create parser for case %s", c.name) {
				for _, table := range p.TablesWithPrimaryKey {
					assert.Contains(t, c.expected, table.String(), "unexpected table found for case %s: %s", c.name, table.String())
				}
			}
		}
	})

	t.Run("should discover relationships", func(t *testing.T) {
		cases := []struct {
			name     string
			mocks    []string
			expected []string
		}{
			{
				name:  "one-to-one",
				mocks: []string{"001_one_to_one/001_tables.sql", "001_one_to_one/002_records.sql"},
				expected: []string{
					"many-to-one | public.orders.[{Name:user_id DataType:integer IsPrimary:false}] -> public.users.[{Name:id DataType:integer IsPrimary:true}]",
					"many-to-one | public.payments.[{Name:order_id DataType:integer IsPrimary:false}] -> public.orders.[{Name:id DataType:integer IsPrimary:true}]",
				},
			},
			{
				name:  "many-to-many",
				mocks: []string{"002_many_to_many/001_tables.sql", "002_many_to_many/002_records.sql"},
				expected: []string{
					"one-to-one | public.user_orders.[{Name:user_id DataType:integer IsPrimary:true}] -> public.users.[{Name:id DataType:integer IsPrimary:true}]",
					"one-to-one | public.user_orders.[{Name:order_id DataType:integer IsPrimary:true}] -> public.orders.[{Name:id DataType:integer IsPrimary:true}]",
					"one-to-one | public.order_payments.[{Name:order_id DataType:integer IsPrimary:true}] -> public.orders.[{Name:id DataType:integer IsPrimary:true}]",
					"one-to-one | public.order_payments.[{Name:payment_id DataType:integer IsPrimary:true}] -> public.payments.[{Name:payment_id DataType:integer IsPrimary:true}]",
				},
			},
			{
				name:  "self-referencing",
				mocks: []string{"005_self_referencing/001_tables.sql", "005_self_referencing/002_records.sql"},
				expected: []string{
					"many-to-one | public.persons.[{Name:gender_id DataType:integer IsPrimary:false}] -> public.genders.[{Name:gender_id DataType:integer IsPrimary:true}]",
					"self-referencing | public.persons.[{Name:parent_id DataType:integer IsPrimary:false}] -> public.persons.[{Name:person_id DataType:integer IsPrimary:true}]",
				},
			},
		}

		for _, c := range cases {
			_, pgPool := NewPostgresContainer(ctx, t, c.mocks...)
			p, err := parser.NewParser(pgPool, parser.NewParserConfig())
			if assert.NoError(t, err, "failed to create parser for case %s", c.name) {
				for _, rel := range p.Relationships {
					assert.Contains(t, c.expected, rel.String(), "unexpected relationship found for case %s: %s", c.name, rel.String())
				}
			}
		}
	})

	t.Run("should have map of table to pk columns", func(t *testing.T) {
		cases := []struct {
			name     string
			mocks    []string
			expected map[string][]parser.Column
		}{
			{
				name:  "one-to-one",
				mocks: []string{"001_one_to_one/001_tables.sql", "001_one_to_one/002_records.sql"},
				expected: map[string][]parser.Column{
					"public.users":    {parser.Column{Name: "id", DataType: "integer", IsPrimary: true}},
					"public.orders":   {parser.Column{Name: "id", DataType: "integer", IsPrimary: true}},
					"public.payments": {parser.Column{Name: "id", DataType: "integer", IsPrimary: true}},
				},
			},
			{
				name:  "many-to-many",
				mocks: []string{"002_many_to_many/001_tables.sql", "002_many_to_many/002_records.sql"},
				expected: map[string][]parser.Column{
					"public.users":    {parser.Column{Name: "id", DataType: "integer", IsPrimary: true}},
					"public.orders":   {parser.Column{Name: "id", DataType: "integer", IsPrimary: true}},
					"public.payments": {parser.Column{Name: "payment_id", DataType: "integer", IsPrimary: true}},
					"public.user_orders": {
						parser.Column{Name: "user_id", DataType: "integer", IsPrimary: true},
						parser.Column{Name: "order_id", DataType: "integer", IsPrimary: true},
					},
					"public.order_payments": {
						parser.Column{Name: "order_id", DataType: "integer", IsPrimary: true},
						parser.Column{Name: "payment_id", DataType: "integer", IsPrimary: true},
					},
				},
			},
			{
				name:  "self-referencing",
				mocks: []string{"005_self_referencing/001_tables.sql", "005_self_referencing/002_records.sql"},
				expected: map[string][]parser.Column{
					"public.genders": {parser.Column{Name: "gender_id", DataType: "integer", IsPrimary: true}},
					"public.persons": {parser.Column{Name: "person_id", DataType: "integer", IsPrimary: true}},
				},
			},
		}

		for _, c := range cases {
			_, pgPool := NewPostgresContainer(ctx, t, c.mocks...)
			p, err := parser.NewParser(pgPool, parser.NewParserConfig())
			if assert.NoError(t, err, "failed to create parser for case %s", c.name) {
				for tableName, columns := range p.TableToPKColumnsMap {
					assert.Equal(t, c.expected[tableName], columns, "unexpected primary key columns for case %s table %s", c.name, tableName)
				}
			}
		}
	})

	t.Run("should fetch record", func(t *testing.T) {
		cases := []struct {
			name   string
			mocks  []string
			checks []struct {
				table  parser.Table      // this is the table being checked
				pk     parser.PrimaryKey // this is the primary key of the table being checked
				record string            // this is the expected record for the table being checked
			}
		}{
			{
				name:  "one-to-one",
				mocks: []string{"001_one_to_one/001_tables.sql", "001_one_to_one/002_records.sql"},
				checks: []struct {
					table  parser.Table
					pk     parser.PrimaryKey
					record string
				}{
					{
						table: parser.Table{Name: "users", Schema: "public", Columns: []parser.Column{
							{Name: "id", DataType: "integer", IsPrimary: true},
							{Name: "name", DataType: "character varying", IsPrimary: false},
						}},
						pk:     parser.PrimaryKey{Columns: []parser.Column{{Name: "id", DataType: "integer", IsPrimary: true}}, Values: []interface{}{1}},
						record: "{public.users [{Name:id DataType:integer IsPrimary:true} {Name:name DataType:character varying IsPrimary:false}] [1 John Doe]}",
					},
					{
						table: parser.Table{Name: "orders", Schema: "public", Columns: []parser.Column{
							{Name: "id", DataType: "integer", IsPrimary: true},
							{Name: "user_id", DataType: "integer", IsPrimary: false},
							{Name: "amount", DataType: "numeric", IsPrimary: false},
						}},
						pk:     parser.PrimaryKey{Columns: []parser.Column{{Name: "id", DataType: "integer", IsPrimary: true}}, Values: []interface{}{1}},
						record: "{public.orders [{Name:id DataType:integer IsPrimary:true} {Name:user_id DataType:integer IsPrimary:false} {Name:amount DataType:numeric IsPrimary:false}] [1 1 {Int:+9999 Exp:-2 NaN:false InfinityModifier:finite Valid:true}]}",
					},
					{
						table: parser.Table{Name: "payments", Schema: "public", Columns: []parser.Column{
							{Name: "id", DataType: "integer", IsPrimary: true},
							{Name: "order_id", DataType: "integer", IsPrimary: false},
							{Name: "amount", DataType: "numeric", IsPrimary: false},
						}},
						pk:     parser.PrimaryKey{Columns: []parser.Column{{Name: "id", DataType: "integer", IsPrimary: true}}, Values: []interface{}{1}},
						record: "{public.payments [{Name:id DataType:integer IsPrimary:true} {Name:order_id DataType:integer IsPrimary:false} {Name:amount DataType:numeric IsPrimary:false}] [1 1 {Int:+9999 Exp:-2 NaN:false InfinityModifier:finite Valid:true}]}",
					},
				},
			},
		}

		for _, c := range cases {
			_, pgPool := NewPostgresContainer(ctx, t, c.mocks...)
			p, err := parser.NewParser(pgPool, parser.NewParserConfig())
			if assert.NoError(t, err, "failed to create parser for case %s", c.name) {
				for _, check := range c.checks {
					record, err := p.FetchRecord(ctx, check.table, check.pk)
					if assert.NoError(t, err, "failed to fetch record for case %s", c.name) {
						assert.True(t, record.Equal(record), "unexpected record for case %s: %+v, should be: %+v", c.name, record, check.record)
					}
				}
			}
		}
	})

	t.Run("should build records graph", func(t *testing.T) {
		cases := []struct {
			name    string
			mocks   []string
			schemas []string
			checks  []struct {
				table   parser.Table      // this is the table being checked
				pk      parser.PrimaryKey // this is the primary key of the table being checked
				results []parser.Record   // this is the expected result of the query
			}
		}{
			{
				name:    "one-to-one",
				mocks:   []string{"001_one_to_one/001_tables.sql", "001_one_to_one/002_records.sql"},
				schemas: []string{"public"},
				checks: []struct {
					table   parser.Table
					pk      parser.PrimaryKey
					results []parser.Record
				}{
					{
						// we search in this table
						table: parser.Table{
							Name:   "orders",
							Schema: "public",
						},
						// we search for this primary key
						pk: parser.PrimaryKey{
							Columns: []parser.Column{
								{Name: "id", DataType: "integer", IsPrimary: true},
							},
							Values: []interface{}{1},
						},
						// and we expected to get the following records
						results: []parser.Record{
							{
								Table:   parser.Table{Name: "users", Schema: "public"},
								Columns: []parser.Column{{Name: "id"}, {Name: "name"}},
								Values:  []interface{}{int32(1), "John Doe"},
							},
							{
								Table:   parser.Table{Name: "orders", Schema: "public"},
								Columns: []parser.Column{{Name: "id"}, {Name: "user_id"}, {Name: "amount"}},
								Values:  []interface{}{int32(1), int32(1), pgtype.Numeric{Int: big.NewInt(9999), Exp: -2, Valid: true}},
							},
							{
								Table:   parser.Table{Name: "payments", Schema: "public"},
								Columns: []parser.Column{{Name: "id"}, {Name: "order_id"}, {Name: "amount"}},
								Values:  []interface{}{int32(1), int32(1), pgtype.Numeric{Int: big.NewInt(9999), Exp: -2, Valid: true}},
							},
						},
					},
				},
			},
			{
				name:    "many-to-many",
				mocks:   []string{"002_many_to_many/001_tables.sql", "002_many_to_many/002_records.sql"},
				schemas: []string{"public"},
				checks: []struct {
					table   parser.Table
					pk      parser.PrimaryKey
					results []parser.Record
				}{
					{
						// we search in this table
						table: parser.Table{
							Name:   "orders",
							Schema: "public",
						},
						// we search for this primary key
						pk: parser.PrimaryKey{
							Columns: []parser.Column{
								{Name: "id", DataType: "integer", IsPrimary: true},
							},
							Values: []interface{}{1},
						},
						// and we expected to get the following records
						results: []parser.Record{
							{
								Table:   parser.Table{Name: "orders", Schema: "public"},
								Columns: []parser.Column{{Name: "id"}, {Name: "amount"}},
								Values:  []interface{}{int32(1), pgtype.Numeric{Int: big.NewInt(10050), Exp: -2, Valid: true}},
							},
							{
								Table:   parser.Table{Name: "users", Schema: "public"},
								Columns: []parser.Column{{Name: "id"}, {Name: "name"}},
								Values:  []interface{}{int32(1), "John Doe"},
							},
							{
								Table:   parser.Table{Name: "user_orders", Schema: "public"},
								Columns: []parser.Column{{Name: "user_id"}, {Name: "order_id"}},
								Values:  []interface{}{int32(1), int32(1)},
							},
							{
								Table:   parser.Table{Name: "payments", Schema: "public"},
								Columns: []parser.Column{{Name: "payment_id"}, {Name: "amount"}},
								Values:  []interface{}{int32(1), pgtype.Numeric{Int: big.NewInt(5025), Exp: -2, Valid: true}},
							},
							{
								Table:   parser.Table{Name: "order_payments", Schema: "public"},
								Columns: []parser.Column{{Name: "order_id"}, {Name: "payment_id"}},
								Values:  []interface{}{int32(1), int32(1)},
							},
							{
								Table:   parser.Table{Name: "payments", Schema: "public"},
								Columns: []parser.Column{{Name: "payment_id"}, {Name: "amount"}},
								Values:  []interface{}{int32(2), pgtype.Numeric{Int: big.NewInt(5025), Exp: -2, Valid: true}},
							},
							{
								Table:   parser.Table{Name: "order_payments", Schema: "public"},
								Columns: []parser.Column{{Name: "order_id"}, {Name: "payment_id"}},
								Values:  []interface{}{int32(1), int32(2)},
							},
						},
					},
				},
			},
			{
				name:    "self-referencing",
				mocks:   []string{"005_self_referencing/001_tables.sql", "005_self_referencing/002_records.sql"},
				schemas: []string{"public"},
				checks: []struct {
					table   parser.Table
					pk      parser.PrimaryKey
					results []parser.Record
				}{
					{
						// we search in this table
						table: parser.Table{
							Name:   "persons",
							Schema: "public",
							Columns: []parser.Column{
								{Name: "person_id", DataType: "integer", IsPrimary: true},
								{Name: "first_name", DataType: "character varying", IsPrimary: false},
								{Name: "gender_id", DataType: "integer", IsPrimary: false},
								{Name: "parent_id", DataType: "integer", IsPrimary: false},
							},
						},
						// we search for this primary key
						pk: parser.PrimaryKey{
							Columns: []parser.Column{
								{Name: "person_id", DataType: "integer", IsPrimary: true},
							},
							Values: []interface{}{3},
						},
						// and we expected to get the following records
						results: []parser.Record{
							{
								Table:   parser.Table{Name: "genders", Schema: "public"},
								Columns: []parser.Column{{Name: "gender_id"}, {Name: "gender_name"}},
								Values:  []interface{}{int32(1), "Male"},
							},
							{
								Table:   parser.Table{Name: "persons", Schema: "public"},
								Columns: []parser.Column{{Name: "person_id"}, {Name: "first_name"}, {Name: "gender_id"}, {Name: "parent_id"}},
								Values:  []interface{}{int32(1), "John", int32(1), nil},
							},
							{
								Table:   parser.Table{Name: "persons", Schema: "public"},
								Columns: []parser.Column{{Name: "person_id"}, {Name: "first_name"}, {Name: "gender_id"}, {Name: "parent_id"}},
								Values:  []interface{}{int32(3), "James", int32(1), int32(1)},
							},
						},
					},
				},
			},
		}

		for _, c := range cases {
			_, pgPool := NewPostgresContainer(ctx, t, c.mocks...)
			p, err := parser.NewParser(pgPool, parser.NewParserConfig(parser.WithSchemas(c.schemas)))
			if assert.NoError(t, err, "failed to create parser for case %s", c.name) {
				for _, check := range c.checks {
					records, err := p.BuildGraph(ctx, check.table, check.pk)
					if assert.NoError(t, err, "failed to build graph for case %s", c.name) {
						assert.Len(t, records, len(check.results), "failed to build graph for case %s", c.name)
						for i, record := range records {
							if i > len(check.results)-1 {
								t.Fatalf("not enough expected results for case %s, received %d, but expected %d", c.name, len(records), len(check.results))
								break
							}
							assert.True(t, record.Equal(check.results[i]), "unexpected record idx %d for case %s: %+v, should be: %+v", i, c.name, record, check.results[i])
						}
					}
				}
			}
		}
	})

	t.Run("should extract records graph", func(t *testing.T) {
		cases := []struct {
			name        string
			tableMocks  []string
			recordMocks []string
			schemas     []string
			checks      []struct {
				table parser.Table      // this is the table being checked
				pk    parser.PrimaryKey // this is the primary key of the table being checked
				sql   string            // this is what we expect in return from ExtractGraph
				error error             // this is error we expect to receive from ExtractGraph
			}
		}{
			// {
			// 	name:        "simple circular dependencies",
			// 	tableMocks:  []string{"003_circular_simple/001_tables.sql"},
			// 	recordMocks: []string{"003_circular_simple/002_records.sql"},
			// 	schemas:     []string{"example"},
			// 	checks: []struct {
			// 		table parser.Table
			// 		pk    parser.PrimaryKey
			// 		sql   string
			// 		error error
			// 	}{
			// 		{
			// 			table: parser.Table{
			// 				Name:   "persons",
			// 				Schema: "example",
			// 			},
			// 			pk: parser.PrimaryKey{
			// 				Columns: []parser.Column{
			// 					{Name: "person_id", DataType: "integer", IsPrimary: true},
			// 				},
			// 				Values: []interface{}{1},
			// 			},
			// 			sql: "INSERT INTO example.countries (country_id, code) VALUES (1, 'USA');\n" +
			// 				"INSERT INTO example.cars (car_id, make, country_of_origin_id) VALUES (1, 'Ford', 1);\n" +
			// 				"INSERT INTO example.persons (person_id, first_name, country_of_origin_id, car_id) VALUES (1, 'John', 1, 1);\n",
			// 			error: nil,
			// 		},
			// 	},
			// },
			// {
			// 	name:        "circular dependencies with endless loop",
			// 	tableMocks:  []string{"004_circular_loop/001_tables.sql"},
			// 	recordMocks: []string{"004_circular_loop/002_records.sql"},
			// 	schemas:     []string{"example"},
			// 	checks: []struct {
			// 		table parser.Table
			// 		pk    parser.PrimaryKey
			// 		sql   string
			// 		error error
			// 	}{
			// 		{
			// 			table: parser.Table{
			// 				Name:   "departments",
			// 				Schema: "example",
			// 			},
			// 			pk: parser.PrimaryKey{
			// 				Columns: []parser.Column{
			// 					{Name: "department_id", DataType: "integer", IsPrimary: true},
			// 				},
			// 				Values: []interface{}{5},
			// 			},
			// 			error: nil,
			// 		},
			// 	},
			// },
			{
				name:        "deduplication",
				tableMocks:  []string{"006_deduplication/001_tables.sql"},
				recordMocks: []string{"006_deduplication/002_records.sql"},
				schemas:     []string{"public"},
				checks: []struct {
					table parser.Table
					pk    parser.PrimaryKey
					sql   string
					error error
				}{
					{
						table: parser.Table{
							Name:   "tasks",
							Schema: "public",
						},
						pk: parser.PrimaryKey{
							Columns: []parser.Column{
								{Name: "task_id", DataType: "integer", IsPrimary: true},
							},
							Values: []interface{}{1},
						},
						sql: func() string {
							return ""
						}(),
						error: nil,
					},
				},
			},
			// {
			// 	name:        "data types",
			// 	tableMocks:  []string{"100_data_types/001_tables.sql"},
			// 	recordMocks: []string{"100_data_types/002_records.sql"},
			// 	schemas:     []string{"public"},
			// 	checks: []struct {
			// 		table parser.Table
			// 		pk    parser.PrimaryKey
			// 		sql   string
			// 		error error
			// 	}{
			// 		{
			// 			table: parser.Table{
			// 				Name:   "persons",
			// 				Schema: "public",
			// 			},
			// 			pk: parser.PrimaryKey{
			// 				Columns: []parser.Column{
			// 					{Name: "id", DataType: "integer", IsPrimary: true},
			// 				},
			// 				Values: []interface{}{1},
			// 			},
			// 			sql: func() string {
			// 				t, _ := time.Parse(time.RFC3339Nano, "2025-04-10T12:55:25.034657+03:00")
			// 				localTime := t.Local()
			// 				localStr := localTime.Format(time.RFC3339Nano)
			// 				return "INSERT INTO public.persons (id, name) VALUES (1, 'John Doe');\n" +
			// 					"INSERT INTO public.cars (id, owner_id, make, model, production_year, price, mileage, engine_capacity, weight, is_electric, purchase_date, maintenance_time, registered_at, features, car_numbers, body_color, fuel_capacity, zero_to_60_seconds, previous_owners, warranty_duration, car_image, color_codes, license_plate, ip_address, mac_address, serial_bits, search_vector, geometric_data, uuid, constraint_code) VALUES (1, 1, 'Toyota', 'Camry', 2020, 25000.5, 15000, 2.5, 1560.75, false, '2021-03-15T00:00:00Z', '08:30:00', '" + localStr + "', '{\"sunroof\": false, \"navigation\": true}', '{ABC-123,XYZ-789}', 'blue', NULL, '00:00:06.2', '{}', NULL, NULL, NULL, '192.168.1.0/24', '192.168.1.1/32', '08:00:2b:01:02:03', '101010', '''brown'':2 ''fox'':3 ''quick'':1', '(12.34,56.78)', '77764b84-d905-4519-b3cb-222f6ca0d09e', 123);\n"
			// 			}(),
			// 			error: nil,
			// 		},
			// 	},
			// },
		}

		for _, c := range cases {
			_, pgPoolTest := NewPostgresContainer(ctx, t, c.tableMocks...)
			mocks := c.tableMocks
			mocks = append(mocks, c.recordMocks...)
			_, pgPool := NewPostgresContainer(ctx, t, mocks...)
			p, err := parser.NewParser(pgPool, parser.NewParserConfig(parser.WithSchemas(c.schemas)))
			if assert.NoError(t, err, "failed to create parser for case %s", c.name) {
				for _, check := range c.checks {
					sql, err := p.ExtractGraph(ctx, check.table, check.pk)
					if check.error != nil {
						assert.Errorf(t, err, check.error.Error(), "should have received error")
					} else if check.error == nil && err != nil {
						t.Fatalf("unexpected error: %+v", err)
					} else {
						if assert.NoError(t, err, "failed to extract graph for case %s", c.name) {
							if !assert.True(t, strings.Compare(sql, check.sql) == 0) {
								println("Expected:")
								println(check.sql)
								println("Actual:")
								println(sql)
							}
							// Try to insert generated SQL statements into a freshly prepared database
							_, err = pgPoolTest.Exec(ctx, sql)
							assert.NoError(t, err)
						}
					}
				}
			}
		}
	})

	t.Run("should generate insert statements", func(t *testing.T) {
		_, pgPool := NewPostgresContainer(ctx, t)
		p, _ := parser.NewParser(pgPool, parser.NewParserConfig())
		// Create a fixed time for testing
		testTime, _ := time.Parse(time.RFC3339, "2023-01-02T15:04:05Z")
		// Create a pgtype.Numeric for testing
		numeric := pgtype.Numeric{Int: big.NewInt(12345), Exp: -2, Valid: true}

		cases := []struct {
			name          string
			records       []parser.Record
			expectedSQL   string
			expectedError error
		}{
			{
				name: "Basic insert",
				records: []parser.Record{
					{
						Table:   parser.Table{Name: "users", Schema: "public"},
						Columns: []parser.Column{{Name: "id"}, {Name: "name"}},
						Values:  []interface{}{1, "John"},
					},
				},
				expectedSQL:   "INSERT INTO public.users (id, name) VALUES (1, 'John');\n",
				expectedError: nil,
			},
			{
				name: "Multiple records",
				records: []parser.Record{
					{
						Table:   parser.Table{Name: "users", Schema: "public"},
						Columns: []parser.Column{{Name: "id"}, {Name: "name"}},
						Values:  []interface{}{1, "John"},
					},
					{
						Table:   parser.Table{Name: "users", Schema: "public"},
						Columns: []parser.Column{{Name: "id"}, {Name: "name"}},
						Values:  []interface{}{2, "Jane"},
					},
				},
				expectedSQL:   "INSERT INTO public.users (id, name) VALUES (1, 'John');\nINSERT INTO public.users (id, name) VALUES (2, 'Jane');\n",
				expectedError: nil,
			},
			{
				name: "Different tables",
				records: []parser.Record{
					{
						Table:   parser.Table{Name: "users", Schema: "public"},
						Columns: []parser.Column{{Name: "id"}, {Name: "name"}},
						Values:  []interface{}{1, "John"},
					},
					{
						Table:   parser.Table{Name: "products", Schema: "public"},
						Columns: []parser.Column{{Name: "id"}, {Name: "name"}, {Name: "price"}},
						Values:  []interface{}{101, "Widget", 19.99},
					},
				},
				expectedSQL:   "INSERT INTO public.users (id, name) VALUES (1, 'John');\nINSERT INTO public.products (id, name, price) VALUES (101, 'Widget', 19.99);\n",
				expectedError: nil,
			},
			{
				name: "Null values",
				records: []parser.Record{
					{
						Table:   parser.Table{Name: "users", Schema: "public"},
						Columns: []parser.Column{{Name: "id"}, {Name: "name"}, {Name: "email"}},
						Values:  []interface{}{1, "John", nil},
					},
				},
				expectedSQL:   "INSERT INTO public.users (id, name, email) VALUES (1, 'John', NULL);\n",
				expectedError: nil,
			},
			{
				name: "String escaping",
				records: []parser.Record{
					{
						Table:   parser.Table{Name: "quotes", Schema: "public"},
						Columns: []parser.Column{{Name: "id"}, {Name: "quote"}},
						Values:  []interface{}{1, "It's a 'quoted' string"},
					},
				},
				expectedSQL:   "INSERT INTO public.quotes (id, quote) VALUES (1, 'It''s a ''quoted'' string');\n",
				expectedError: nil,
			},
			{
				name: "JSON data",
				records: []parser.Record{
					{
						Table:   parser.Table{Name: "json_data", Schema: "public"},
						Columns: []parser.Column{{Name: "id"}, {Name: "data"}},
						Values:  []interface{}{1, map[string]interface{}{"name": "John", "age": 30}},
					},
				},
				expectedSQL:   "INSERT INTO public.json_data (id, data) VALUES (1, '{\"age\":30,\"name\":\"John\"}');\n",
				expectedError: nil,
			},
			{
				name: "JSON array",
				records: []parser.Record{
					{
						Table:   parser.Table{Name: "json_arrays", Schema: "public"},
						Columns: []parser.Column{{Name: "id"}, {Name: "tags"}},
						Values:  []interface{}{1, []interface{}{"tag1", "tag2", "tag3"}},
					},
				},
				expectedSQL:   "INSERT INTO public.json_arrays (id, tags) VALUES (1, '[\"tag1\",\"tag2\",\"tag3\"]');\n",
				expectedError: nil,
			},
			{
				name: "Integer types",
				records: []parser.Record{
					{
						Table:   parser.Table{Name: "numbers", Schema: "public"},
						Columns: []parser.Column{{Name: "int_val"}, {Name: "int32_val"}, {Name: "int64_val"}},
						Values:  []interface{}{int(10), int32(20), int64(30)},
					},
				},
				expectedSQL:   "INSERT INTO public.numbers (int_val, int32_val, int64_val) VALUES (10, 20, 30);\n",
				expectedError: nil,
			},
			{
				name:          "Empty record set",
				records:       []parser.Record{},
				expectedSQL:   "",
				expectedError: nil,
			},
			{
				name: "Various data types",
				records: []parser.Record{
					{
						Table:   parser.Table{Name: "types", Schema: "public"},
						Columns: []parser.Column{{Name: "int_val"}, {Name: "float_val"}, {Name: "bool_val"}, {Name: "string_val"}, {Name: "bytes_val"}, {Name: "time_val"}, {Name: "numeric_val"}},
						Values:  []interface{}{42, 3.14, true, "text", []byte("bytes"), testTime, numeric},
					},
				},
				expectedSQL:   "INSERT INTO public.types (int_val, float_val, bool_val, string_val, bytes_val, time_val, numeric_val) VALUES (42, 3.14, true, 'text', 'bytes', '2023-01-02T15:04:05Z', 123.45);\n",
				expectedError: nil,
			},
		}

		for _, c := range cases {
			t.Run(c.name, func(t *testing.T) {
				sql, err := p.GenerateInsertStatements(ctx, c.records)
				if assert.NoError(t, err, "error generating insert statements for case %s", c.name) {
					if c.expectedError == nil && err == nil {
						assert.Equal(t, c.expectedSQL, sql, "SQL statements should be equal in case %s", c.name)
						return
					}
					if c.expectedError != nil {
						assert.Equal(t, c.expectedError, err, "errors should be equal in case %s", c.name)
					}
				}
			})
		}
	})
}
