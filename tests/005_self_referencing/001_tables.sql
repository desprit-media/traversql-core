CREATE TABLE genders (
	gender_id SERIAL PRIMARY KEY,
	gender_name VARCHAR(50) NOT NULL
);

-- Create persons table with self-reference
CREATE TABLE persons (
	person_id SERIAL PRIMARY KEY,
	first_name VARCHAR(100) NOT NULL,
	gender_id INTEGER NOT NULL REFERENCES genders(gender_id),
	parent_id INTEGER REFERENCES persons(person_id)
);