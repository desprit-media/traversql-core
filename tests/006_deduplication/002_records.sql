-- Insert into departments (Table C)
INSERT INTO departments (department_name) VALUES 
('Engineering'),
('Marketing'),
('Human Resources'),
('Finance');

-- Insert into projects (Table B)
INSERT INTO projects (project_name, department_id) VALUES 
('Database Migration', 1),
('Website Redesign', 2),
('Employee Training', 3),
('Budget Planning', 4);

-- Insert into tasks (Table A)
INSERT INTO tasks (task_name, project_id, department_id) VALUES 
('Schema Design', 1, 1),
('Data Migration', 1, 1),
('UI Design', 2, 2),
('Content Creation', 2, 2),
('Training Materials', 3, 3),
('Financial Analysis', 4, 4);

-- Insert into employees
INSERT INTO employees (first_name, last_name, department_id) VALUES 
('John', 'Smith', 1),
('Mary', 'Johnson', 2),
('Robert', 'Williams', 3),
('Lisa', 'Brown', 4);

-- Insert into skills
INSERT INTO skills (skill_name, employee_id) VALUES 
('SQL', 1),
('Java', 1),
('Python', 2),
('Graphic Design', 3),
('Public Speaking', 4);