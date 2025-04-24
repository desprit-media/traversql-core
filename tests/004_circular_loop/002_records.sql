-- We need to carefully insert data to handle the circular dependencies

-- First, insert departments with NULL manager_id
INSERT INTO example.departments (name, manager_id) VALUES
('Engineering', NULL),
('Marketing', NULL),
('Human Resources', NULL),
('Finance', NULL),
('Research', NULL);

-- Insert users with department_id
INSERT INTO example.users (username, department_id) VALUES
('jsmith', 1),
('agarcia', 2),
('bwilson', 3),
('cjohnson', 4),
('dlee', 5);

-- Update departments with manager_id
UPDATE example.departments SET manager_id = 1 WHERE department_id = 1;
UPDATE example.departments SET manager_id = 2 WHERE department_id = 2;
UPDATE example.departments SET manager_id = 3 WHERE department_id = 3;
UPDATE example.departments SET manager_id = 4 WHERE department_id = 4;
UPDATE example.departments SET manager_id = 5 WHERE department_id = 5;

-- Insert projects
INSERT INTO example.projects (name, department_id, lead_id) VALUES
('Website Redesign', 1, 1),
('Summer Campaign', 2, 2),
('Employee Training', 3, 3),
('Budget Analysis', 4, 4),
('New Product Development', 5, 5);

-- Insert tasks (without main task relationship yet)
INSERT INTO example.tasks (name, project_id, parent_task_id) VALUES
('Design mockups', 1, NULL),
('Frontend development', 1, NULL),
('Backend API', 1, NULL),
('Content creation', 2, NULL),
('Media planning', 2, NULL),
('Training materials', 3, NULL),
('Training schedule', 3, NULL),
('Q1 Analysis', 4, NULL),
('Q2 Projections', 4, NULL),
('Market research', 5, NULL),
('Prototype design', 5, NULL);

-- Create subtasks (self-referencing)
UPDATE example.tasks SET parent_task_id = 1 WHERE task_id = 2;
UPDATE example.tasks SET parent_task_id = 1 WHERE task_id = 3;
UPDATE example.tasks SET parent_task_id = 4 WHERE task_id = 5;
UPDATE example.tasks SET parent_task_id = 6 WHERE task_id = 7;
UPDATE example.tasks SET parent_task_id = 10 WHERE task_id = 11;

-- Insert project main tasks (completing another circular dependency)
INSERT INTO example.project_main_tasks (project_id, main_task_id) VALUES
(1, 1),
(2, 4),
(3, 6),
(4, 8),
(5, 10);