-- First, insert countries
INSERT INTO example.countries (code) VALUES
('USA'),
('DEU'),
('JPN'),
('KOR'),
('ITA');

-- Insert cars with country of origin
INSERT INTO example.cars (make, country_of_origin_id) VALUES
('Ford', 1),
('Toyota', 3),
('BMW', 2),
('Hyundai', 4),
('Ferrari', 5),
('Honda', 3),
('Audi', 2);

-- Insert persons with country of origin and car
INSERT INTO example.persons (first_name, country_of_origin_id, car_id) VALUES
('John', 1, 1),    -- American with Ford
('Maria', 1, 2),   -- American with Toyota
('Hans', 2, 3),    -- German with BMW
('Yuki', 3, 6),    -- Japanese with Honda
('Min', 4, 4),     -- Korean with Hyundai
('Sofia', 5, 5),   -- Italian with Ferrari
('James', 1, NULL) -- American with no car
;
