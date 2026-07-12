-- name: CreateCategory :one
INSERT INTO categories(id,name,type,status) VALUES($1,$2,$3,$4) RETURNING *;
-- name: CategoryByID :one
SELECT * FROM categories WHERE id=$1;
-- name: ListCategories :many
SELECT * FROM categories WHERE ($1::text='' OR type=$1) ORDER BY name LIMIT $2 OFFSET $3;
-- name: CountCategories :one
SELECT count(*) FROM categories WHERE ($1::text='' OR type=$1);
-- name: UpdateCategory :one
UPDATE categories SET name=$2,type=$3,status=$4,updated_at=now() WHERE id=$1 RETURNING *;
-- name: DeleteCategory :execrows
DELETE FROM categories WHERE id=$1;

-- name: CreateEmissionFactor :one
INSERT INTO emission_factors(id,name,category_id,unit,kgco2_per_unit,status) VALUES($1,$2,$3,$4,$5,$6) RETURNING *;
-- name: EmissionFactorByID :one
SELECT * FROM emission_factors WHERE id=$1;
-- name: ListEmissionFactors :many
SELECT * FROM emission_factors WHERE ($1::uuid IS NULL OR category_id=$1) ORDER BY name LIMIT $2 OFFSET $3;
-- name: CountEmissionFactors :one
SELECT count(*) FROM emission_factors WHERE ($1::uuid IS NULL OR category_id=$1);
-- name: UpdateEmissionFactor :one
UPDATE emission_factors SET name=$2,category_id=$3,unit=$4,kgco2_per_unit=$5,status=$6,updated_at=now() WHERE id=$1 RETURNING *;
-- name: DeleteEmissionFactor :execrows
DELETE FROM emission_factors WHERE id=$1;

-- name: CreateProductProfile :one
INSERT INTO product_esg_profiles(id,product,attributes,emission_factor_id) VALUES($1,$2,$3,$4) RETURNING *;
-- name: ProductProfileByID :one
SELECT * FROM product_esg_profiles WHERE id=$1;
-- name: ListProductProfiles :many
SELECT * FROM product_esg_profiles ORDER BY product LIMIT $1 OFFSET $2;
-- name: CountProductProfiles :one
SELECT count(*) FROM product_esg_profiles;
-- name: UpdateProductProfile :one
UPDATE product_esg_profiles SET product=$2,attributes=$3,emission_factor_id=$4,updated_at=now() WHERE id=$1 RETURNING *;
-- name: DeleteProductProfile :execrows
DELETE FROM product_esg_profiles WHERE id=$1;

-- name: CreatePolicy :one
INSERT INTO esg_policies(id,title,body,version,effective_date) VALUES($1,$2,$3,$4,$5) RETURNING *;
-- name: PolicyByID :one
SELECT * FROM esg_policies WHERE id=$1;
-- name: ListPolicies :many
SELECT * FROM esg_policies ORDER BY title LIMIT $1 OFFSET $2;
-- name: CountPolicies :one
SELECT count(*) FROM esg_policies;
-- name: UpdatePolicy :one
UPDATE esg_policies SET title=$2,body=$3,version=$4,effective_date=$5,updated_at=now() WHERE id=$1 RETURNING *;
-- name: DeletePolicy :execrows
DELETE FROM esg_policies WHERE id=$1;

-- name: CreateBadge :one
INSERT INTO badges(id,name,description,icon,unlock_rule) VALUES($1,$2,$3,$4,$5) RETURNING *;
-- name: BadgeByID :one
SELECT * FROM badges WHERE id=$1;
-- name: ListBadges :many
SELECT * FROM badges ORDER BY name LIMIT $1 OFFSET $2;
-- name: CountBadges :one
SELECT count(*) FROM badges;
-- name: UpdateBadge :one
UPDATE badges SET name=$2,description=$3,icon=$4,unlock_rule=$5,updated_at=now() WHERE id=$1 RETURNING *;
-- name: DeleteBadge :execrows
DELETE FROM badges WHERE id=$1;

-- name: CreateReward :one
INSERT INTO rewards(id,name,description,points_required,stock,status) VALUES($1,$2,$3,$4,$5,$6) RETURNING *;
-- name: RewardByID :one
SELECT * FROM rewards WHERE id=$1;
-- name: ListRewards :many
SELECT * FROM rewards ORDER BY name LIMIT $1 OFFSET $2;
-- name: CountRewards :one
SELECT count(*) FROM rewards;
-- name: UpdateReward :one
UPDATE rewards SET name=$2,description=$3,points_required=$4,stock=$5,status=$6,updated_at=now() WHERE id=$1 RETURNING *;
-- name: DeleteReward :execrows
DELETE FROM rewards WHERE id=$1;

-- name: GetESGConfig :one
SELECT * FROM esg_config WHERE singleton=TRUE;
-- name: UpdateESGConfig :one
UPDATE esg_config SET auto_emission_calc=$1,require_csr_evidence=$2,auto_award_badges=$3,
 notify_compliance_email=$4,weight_env=$5,weight_social=$6,weight_gov=$7,updated_at=now()
WHERE singleton=TRUE RETURNING *;

-- name: ListNotificationPreferences :many
SELECT * FROM notification_preferences ORDER BY event_type;
-- name: UpsertNotificationPreference :one
INSERT INTO notification_preferences(event_type,in_app_enabled,email_enabled) VALUES($1,$2,$3)
ON CONFLICT(event_type) DO UPDATE SET in_app_enabled=EXCLUDED.in_app_enabled,email_enabled=EXCLUDED.email_enabled,updated_at=now()
RETURNING *;

-- name: ListEmployees :many
SELECT * FROM users ORDER BY name LIMIT $1 OFFSET $2;
-- name: CountEmployees :one
SELECT count(*) FROM users;
-- name: UpdateEmployee :one
UPDATE users SET name=$2,email=$3,role=$4,department_id=$5,status=$6,updated_at=now() WHERE id=$1 RETURNING *;
-- name: DeactivateEmployee :execrows
UPDATE users SET status='inactive',updated_at=now() WHERE id=$1;
