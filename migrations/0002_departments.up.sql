CREATE TABLE departments (
  id UUID PRIMARY KEY,
  name TEXT NOT NULL,
  code TEXT UNIQUE NOT NULL,
  head_id UUID REFERENCES users(id) ON DELETE SET NULL,
  parent_id UUID REFERENCES departments(id) ON DELETE RESTRICT,
  employee_count INT NOT NULL DEFAULT 0 CHECK (employee_count >= 0),
  status TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'inactive')),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

ALTER TABLE users ADD CONSTRAINT fk_users_department
  FOREIGN KEY (department_id) REFERENCES departments(id) ON DELETE SET NULL;

CREATE INDEX idx_departments_parent_id ON departments(parent_id);
CREATE INDEX idx_departments_head_id ON departments(head_id);
CREATE INDEX idx_users_department_id ON users(department_id);
