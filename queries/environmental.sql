-- name: CreateCarbonTransaction :one
INSERT INTO carbon_transactions(id,department_id,source,quantity,emission_factor_id,factor_value,computed_co2,txn_date,evidence_url,status,created_at)
VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11) RETURNING *;

-- name: CarbonTransactionByID :one
SELECT * FROM carbon_transactions WHERE id=$1;

-- name: VerifyCarbonTransaction :one
UPDATE carbon_transactions SET computed_co2=$2,status='verified',verified_by=$3,verified_at=$4
WHERE id=$1 AND status='draft' RETURNING *;

-- name: ListCarbonTransactions :many
SELECT * FROM carbon_transactions
WHERE ($1::uuid IS NULL OR department_id=$1)
  AND ($2::date IS NULL OR txn_date >= $2)
  AND ($3::date IS NULL OR txn_date <= $3)
  AND ($4::text = '' OR source=$4)
  AND ($5::text = '' OR status=$5)
ORDER BY txn_date DESC, created_at DESC LIMIT $6 OFFSET $7;

-- name: CountCarbonTransactions :one
SELECT count(*) FROM carbon_transactions
WHERE ($1::uuid IS NULL OR department_id=$1)
  AND ($2::date IS NULL OR txn_date >= $2)
  AND ($3::date IS NULL OR txn_date <= $3)
  AND ($4::text = '' OR source=$4)
  AND ($5::text = '' OR status=$5);

-- name: CarbonSummary :many
SELECT source, COALESCE(sum(computed_co2),0)::numeric(14,3) AS co2
FROM carbon_transactions
WHERE status='verified' AND ($1::uuid IS NULL OR department_id=$1) AND txn_date BETWEEN $2 AND $3
GROUP BY source ORDER BY source;

-- name: ActiveEmissionFactor :one
SELECT unit,kgco2_per_unit,status FROM emission_factors WHERE id=$1;

-- name: DepartmentExists :one
SELECT EXISTS(SELECT 1 FROM departments WHERE id=$1 AND status='active');

-- name: IsDepartmentHead :one
SELECT EXISTS(SELECT 1 FROM departments WHERE id=$1 AND head_id=$2 AND status='active');

-- name: CreateEnvironmentalGoal :one
INSERT INTO environmental_goals(id,name,department_id,target_co2,current_co2,deadline,status,created_at,updated_at)
VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9) RETURNING *;

-- name: EnvironmentalGoalByID :one
SELECT * FROM environmental_goals WHERE id=$1;

-- name: UpdateEnvironmentalGoal :one
UPDATE environmental_goals SET name=$2,target_co2=$3,current_co2=$4,deadline=$5,status=$6,updated_at=$7
WHERE id=$1 RETURNING *;

-- name: ListEnvironmentalGoals :many
SELECT * FROM environmental_goals WHERE ($1::uuid IS NULL OR department_id=$1)
ORDER BY deadline,name LIMIT $2 OFFSET $3;

-- name: CountEnvironmentalGoals :one
SELECT count(*) FROM environmental_goals WHERE ($1::uuid IS NULL OR department_id=$1);

-- name: GoalsForDepartment :many
SELECT * FROM environmental_goals WHERE department_id=$1 ORDER BY deadline;

-- name: VerifiedEmissionsThrough :one
SELECT COALESCE(sum(computed_co2),0)::numeric(14,3) FROM carbon_transactions
WHERE status='verified' AND department_id=$1 AND txn_date <= $2;
