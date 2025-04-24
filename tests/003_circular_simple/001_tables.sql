-- Create schema
CREATE SCHEMA IF NOT EXISTS example;

-- Countries table
CREATE TABLE example.countries (
    country_id SERIAL PRIMARY KEY,
    code VARCHAR(3) NOT NULL
);

-- Cars table with country of origin
CREATE TABLE example.cars (
    car_id SERIAL PRIMARY KEY,
    make VARCHAR(50) NOT NULL,
    country_of_origin_id INT NOT NULL,
    CONSTRAINT fk_car_country FOREIGN KEY (country_of_origin_id)
        REFERENCES example.countries (country_id)
);

-- Persons table with country of origin and car
CREATE TABLE example.persons (
    person_id SERIAL PRIMARY KEY,
    first_name VARCHAR(50) NOT NULL,
    country_of_origin_id INT NOT NULL,
    car_id INT NULL,
    CONSTRAINT fk_person_country FOREIGN KEY (country_of_origin_id)
        REFERENCES example.countries (country_id),
    CONSTRAINT fk_person_car FOREIGN KEY (car_id)
        REFERENCES example.cars (car_id)
);