INSERT INTO genders (gender_id, gender_name) VALUES
(1, 'Male'),
(2, 'Female');

-- First, insert people without parents
INSERT INTO persons (person_id, first_name, gender_id) VALUES
(1, 'John', 1),
(2, 'Mary', 2);

-- Then, insert people with parents (using the IDs generated above)
INSERT INTO persons (person_id, first_name, gender_id, parent_id) VALUES
(3, 'James', 1, 1),
(4, 'Sarah', 2, 1),
(5, 'Michael', 1, 2),
(6, 'Emily', 2, 2);