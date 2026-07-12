-- name: CreateDepartment :one
INSERT INTO departments (id, name, code, head_id, parent_id, employee_count, status)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: DepartmentByID :one
SELECT * FROM departments WHERE id = $1;

-- name: DepartmentCodeExists :one
SELECT EXISTS(SELECT 1 FROM departments WHERE code = $1 AND id <> $2);

-- name: ListDepartments :many
SELECT * FROM departments ORDER BY name LIMIT $1 OFFSET $2;

-- name: CountDepartments :one
SELECT count(*) FROM departments;

-- name: UpdateDepartment :one
UPDATE departments SET name=$2, code=$3, head_id=$4, parent_id=$5,
  employee_count=$6, status=$7, updated_at=now()
WHERE id=$1 RETURNING *;

-- name: DeactivateDepartment :one
UPDATE departments SET status='inactive', updated_at=now()
WHERE id=$1 RETURNING *;

-- name: EligibleDepartmentHead :one
SELECT EXISTS(SELECT 1 FROM users WHERE id=$1 AND role IN ('dept_head','admin'));
