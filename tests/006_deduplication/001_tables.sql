-- Create Table C (top of the hierarchy)
CREATE TABLE departments (
    department_id SERIAL PRIMARY KEY,
    department_name VARCHAR(100) NOT NULL
);

-- Create Table B (depends on Table C)
CREATE TABLE projects (
    project_id SERIAL PRIMARY KEY,
    project_name VARCHAR(100) NOT NULL,
    department_id INTEGER REFERENCES departments(department_id)
);

-- Create Table A (depends on both Table B and Table C)
CREATE TABLE tasks (
    task_id SERIAL PRIMARY KEY,
    task_name VARCHAR(100) NOT NULL,
    department_id INTEGER REFERENCES departments(department_id),
    project_id INTEGER REFERENCES projects(project_id)
);

-- Create two additional tables to reach the required five tables
CREATE TABLE employees (
    employee_id SERIAL PRIMARY KEY,
    first_name VARCHAR(50) NOT NULL,
    last_name VARCHAR(50) NOT NULL,
    department_id INTEGER REFERENCES departments(department_id)
);

CREATE TABLE skills (
    skill_id SERIAL PRIMARY KEY,
    skill_name VARCHAR(100) NOT NULL,
    employee_id INTEGER REFERENCES employees(employee_id)
);