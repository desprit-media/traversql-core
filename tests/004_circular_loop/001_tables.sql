-- Create schema
CREATE SCHEMA IF NOT EXISTS example;

-- Users table
CREATE TABLE example.users (
    user_id SERIAL PRIMARY KEY,
    username VARCHAR(50) NOT NULL,
    department_id INT NULL
);

-- Departments table (circular with users - department has manager who is a user)
CREATE TABLE example.departments (
    department_id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    manager_id INT NULL,
    CONSTRAINT fk_department_manager FOREIGN KEY (manager_id) 
        REFERENCES example.users (user_id)
);

-- Add foreign key to users table (completing the circular dependency)
ALTER TABLE example.users 
ADD CONSTRAINT fk_user_department FOREIGN KEY (department_id) 
    REFERENCES example.departments (department_id);

-- Projects table
CREATE TABLE example.projects (
    project_id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    department_id INT NOT NULL,
    lead_id INT NULL,
    CONSTRAINT fk_project_department FOREIGN KEY (department_id)
        REFERENCES example.departments (department_id),
    CONSTRAINT fk_project_lead FOREIGN KEY (lead_id)
        REFERENCES example.users (user_id)
);

-- Tasks table (circular with projects - tasks belong to projects but projects have a "main task")
CREATE TABLE example.tasks (
    task_id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    project_id INT NOT NULL,
    parent_task_id INT NULL,
    CONSTRAINT fk_task_project FOREIGN KEY (project_id)
        REFERENCES example.projects (project_id),
    CONSTRAINT fk_task_parent FOREIGN KEY (parent_task_id)
        REFERENCES example.tasks (task_id)
);

-- Project_main_task table (creates another circular dependency)
CREATE TABLE example.project_main_tasks (
    project_id INT PRIMARY KEY,
    main_task_id INT NOT NULL,
    CONSTRAINT fk_main_task_project FOREIGN KEY (project_id)
        REFERENCES example.projects (project_id),
    CONSTRAINT fk_main_task_task FOREIGN KEY (main_task_id)
        REFERENCES example.tasks (task_id)
);