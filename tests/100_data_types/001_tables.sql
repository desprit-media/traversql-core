-- Create enum type for demonstration
CREATE TYPE color_enum AS ENUM ('red', 'green', 'blue');

-- Persons table
CREATE TABLE persons (
    id INTEGER GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
    name VARCHAR(100) NOT NULL
);

-- Cars table with multiple data types
CREATE TABLE cars (
    id INTEGER GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
    owner_id INTEGER NOT NULL REFERENCES persons(id) ON DELETE CASCADE,
    make VARCHAR(50) NOT NULL,
    model TEXT NOT NULL,
    production_year SMALLINT,
    price NUMERIC(10,2),
    mileage INTEGER,
    engine_capacity REAL,
    weight DOUBLE PRECISION,
    is_electric BOOLEAN DEFAULT false,
    purchase_date DATE,
    maintenance_time TIME,
    registered_at TIMESTAMP WITH TIME ZONE DEFAULT '2025-04-10T12:55:25.034657+03:00',
    features JSONB,
    car_numbers TEXT[],
    body_color color_enum,
    fuel_capacity DECIMAL(5,1),
    zero_to_60_seconds INTERVAL,
    previous_owners BIGINT[] DEFAULT ARRAY[]::BIGINT[],
    warranty_duration INTERVAL YEAR TO MONTH,
    car_image BYTEA,
    color_codes INT4RANGE,
    license_plate CIDR,
    ip_address INET,
    mac_address MACADDR,
    serial_bits BIT VARYING(8),
    search_vector TSVECTOR,
    geometric_data POINT,
    uuid UUID NOT NULL DEFAULT '77764b84-d905-4519-b3cb-222f6ca0d09e',
    constraint_code INTEGER CHECK (constraint_code BETWEEN 100 AND 999)
);

-- Add index for demonstration
CREATE INDEX cars_make_idx ON cars(make);
