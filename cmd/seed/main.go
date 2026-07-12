package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	platformauth "github.com/siddharthg2309/ecosphere-esg-platform/internal/platform/auth"
)

type department struct{ id, name, code string }
type user struct{ id, name, email, role, department string }

func main() {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()
	tx, err := pool.Begin(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	departments := []department{{"10000000-0000-7000-8000-000000000001", "Manufacturing", "MFG"}, {"10000000-0000-7000-8000-000000000002", "Logistics", "LOG"}, {"10000000-0000-7000-8000-000000000003", "Human Resources", "HR"}, {"10000000-0000-7000-8000-000000000004", "Finance", "FIN"}, {"10000000-0000-7000-8000-000000000005", "Compliance", "CMP"}}
	for _, d := range departments {
		if _, err = tx.Exec(ctx, `INSERT INTO departments(id,name,code,status) VALUES($1,$2,$3,'active') ON CONFLICT(id) DO UPDATE SET name=EXCLUDED.name,code=EXCLUDED.code`, d.id, d.name, d.code); err != nil {
			log.Fatal(err)
		}
	}
	password := env("SEED_ADMIN_PASSWORD", "ChangeMe123!")
	hash, err := platformauth.HashPassword(password)
	if err != nil {
		log.Fatal(err)
	}
	users := []user{{"20000000-0000-7000-8000-000000000001", "Aarav Mehta", env("SEED_ADMIN_EMAIL", "admin@ecosphere.local"), "admin", departments[4].id}, {"20000000-0000-7000-8000-000000000002", "Sneha Nair", "sneha@ecosphere.local", "dept_head", departments[0].id}, {"20000000-0000-7000-8000-000000000003", "Rohan Iyer", "rohan@ecosphere.local", "dept_head", departments[1].id}, {"20000000-0000-7000-8000-000000000004", "Priya Shah", "priya@ecosphere.local", "dept_head", departments[2].id}, {"20000000-0000-7000-8000-000000000005", "Kiran Menon", "kiran@ecosphere.local", "auditor", departments[4].id}}
	for i := 6; i <= 20; i++ {
		dept := departments[(i-1)%len(departments)]
		users = append(users, user{fmt.Sprintf("20000000-0000-7000-8000-%012d", i), fmt.Sprintf("Demo Employee %02d", i), fmt.Sprintf("employee%02d@ecosphere.local", i), "employee", dept.id})
	}
	for _, u := range users {
		if _, err = tx.Exec(ctx, `INSERT INTO users(id,name,email,password_hash,role,department_id,status) VALUES($1,$2,$3,$4,$5,$6,'active') ON CONFLICT(id) DO UPDATE SET name=EXCLUDED.name,email=EXCLUDED.email,role=EXCLUDED.role,department_id=EXCLUDED.department_id`, u.id, u.name, u.email, hash, u.role, u.department); err != nil {
			log.Fatal(err)
		}
	}
	for i := 0; i < 4; i++ {
		_, _ = tx.Exec(ctx, `UPDATE departments SET head_id=$1 WHERE id=$2`, users[i+1].id, departments[i].id)
	}
	categories := [][4]string{{"30000000-0000-7000-8000-000000000001", "Community Service", "csr_activity", "active"}, {"30000000-0000-7000-8000-000000000002", "Health & Wellness", "csr_activity", "active"}, {"30000000-0000-7000-8000-000000000003", "Energy Saving", "challenge", "active"}, {"30000000-0000-7000-8000-000000000004", "Waste Reduction", "challenge", "active"}}
	for _, v := range categories {
		_, err = tx.Exec(ctx, `INSERT INTO categories(id,name,type,status) VALUES($1,$2,$3,$4) ON CONFLICT(id) DO UPDATE SET name=EXCLUDED.name,type=EXCLUDED.type,status=EXCLUDED.status`, v[0], v[1], v[2], v[3])
		if err != nil {
			log.Fatal(err)
		}
	}
	factors := [][5]string{{"40000000-0000-7000-8000-000000000001", "Diesel", categories[2][0], "litre", "2.6800"}, {"40000000-0000-7000-8000-000000000002", "Petrol", categories[2][0], "litre", "2.3100"}, {"40000000-0000-7000-8000-000000000003", "Grid electricity", categories[2][0], "kWh", "0.7100"}, {"40000000-0000-7000-8000-000000000004", "Natural gas", categories[2][0], "kg", "2.7500"}, {"40000000-0000-7000-8000-000000000005", "Air travel", categories[2][0], "passenger-km", "0.1580"}, {"40000000-0000-7000-8000-000000000006", "Rail travel", categories[2][0], "passenger-km", "0.0410"}, {"40000000-0000-7000-8000-000000000007", "Landfill waste", categories[3][0], "kg", "0.5860"}, {"40000000-0000-7000-8000-000000000008", "Recycled waste", categories[3][0], "kg", "0.0210"}}
	for _, v := range factors {
		_, err = tx.Exec(ctx, `INSERT INTO emission_factors(id,name,category_id,unit,kgco2_per_unit,status) VALUES($1,$2,$3,$4,$5,'active') ON CONFLICT(id) DO UPDATE SET name=EXCLUDED.name,category_id=EXCLUDED.category_id,unit=EXCLUDED.unit,kgco2_per_unit=EXCLUDED.kgco2_per_unit`, v[0], v[1], v[2], v[3], v[4])
		if err != nil {
			log.Fatal(err)
		}
	}
	policies := [][4]string{{"50000000-0000-7000-8000-000000000001", "Environmental Responsibility", "Employees must minimize waste and report environmental incidents.", "2026-01-01"}, {"50000000-0000-7000-8000-000000000002", "Supplier Code of Conduct", "Suppliers must meet EcoSphere environmental and social standards.", "2026-01-01"}, {"50000000-0000-7000-8000-000000000003", "Ethics and Governance", "All employees must follow the company ethics and governance controls.", "2026-01-01"}}
	for _, v := range policies {
		_, err = tx.Exec(ctx, `INSERT INTO esg_policies(id,title,body,version,effective_date) VALUES($1,$2,$3,1,$4) ON CONFLICT(id) DO UPDATE SET title=EXCLUDED.title,body=EXCLUDED.body`, v[0], v[1], v[2], v[3])
		if err != nil {
			log.Fatal(err)
		}
	}
	badges := [][6]string{{"60000000-0000-7000-8000-000000000001", "Green Starter", "Earn 100 XP", "leaf", "xp", "100"}, {"60000000-0000-7000-8000-000000000002", "Eco Champion", "Earn 500 XP", "award", "xp", "500"}, {"60000000-0000-7000-8000-000000000003", "Challenge Builder", "Complete 3 challenges", "target", "challenges", "3"}, {"60000000-0000-7000-8000-000000000004", "Sustainability Leader", "Complete 10 challenges", "trophy", "challenges", "10"}}
	for _, v := range badges {
		_, err = tx.Exec(ctx, `INSERT INTO badges(id,name,description,icon,unlock_rule) VALUES($1,$2,$3,$4,jsonb_build_object('type',$5::text,'value',$6::int)) ON CONFLICT(id) DO UPDATE SET name=EXCLUDED.name,description=EXCLUDED.description,icon=EXCLUDED.icon,unlock_rule=EXCLUDED.unlock_rule`, v[0], v[1], v[2], v[3], v[4], v[5])
		if err != nil {
			log.Fatal(err)
		}
	}
	rewards := [][5]string{{"70000000-0000-7000-8000-000000000001", "Reusable Bottle", "EcoSphere steel bottle", "150", "25"}, {"70000000-0000-7000-8000-000000000002", "Plant a Tree", "Tree planted in your name", "250", "100"}, {"70000000-0000-7000-8000-000000000003", "Gift Card", "Sustainable marketplace gift card", "500", "20"}, {"70000000-0000-7000-8000-000000000004", "Volunteer Day", "One paid volunteer day", "1000", "10"}}
	for _, v := range rewards {
		_, err = tx.Exec(ctx, `INSERT INTO rewards(id,name,description,points_required,stock,status) VALUES($1,$2,$3,$4,$5,'active') ON CONFLICT(id) DO UPDATE SET name=EXCLUDED.name,description=EXCLUDED.description,points_required=EXCLUDED.points_required,stock=EXCLUDED.stock`, v[0], v[1], v[2], v[3], v[4])
		if err != nil {
			log.Fatal(err)
		}
	}
	// Phase 3 — demographics for diversity dashboard
	genders := []string{"woman", "man", "woman", "man", "non_binary"}
	for i, u := range users {
		gender := genders[i%len(genders)]
		isLead := u.role == "dept_head" || u.role == "admin"
		_, _ = tx.Exec(ctx, `UPDATE users SET gender=$2, is_leadership=$3 WHERE id=$1`, u.id, gender, isLead)
	}
	// CSR activities
	csr := [][6]string{
		{"80000000-0000-7000-8000-000000000001", "Tree Plantation", categories[0][0], "Plant 500 saplings at the Riverside reserve.", "50", "true"},
		{"80000000-0000-7000-8000-000000000002", "Blood Donation", categories[1][0], "Quarterly donation camp with Red Cross.", "40", "true"},
		{"80000000-0000-7000-8000-000000000003", "Beach Cleanup", categories[0][0], "Coastal cleanup drive, Sunday morning.", "60", "false"},
		{"80000000-0000-7000-8000-000000000004", "ESG Workshop", categories[1][0], "Lunch-and-learn on sustainability basics.", "30", "false"},
	}
	for _, v := range csr {
		_, err = tx.Exec(ctx, `INSERT INTO csr_activities(id,title,category_id,description,points,evidence_required,status)
			VALUES($1,$2,$3,$4,$5::int,$6::bool,'active')
			ON CONFLICT(id) DO UPDATE SET title=EXCLUDED.title,description=EXCLUDED.description,points=EXCLUDED.points`,
			v[0], v[1], v[2], v[3], v[4], v[5])
		if err != nil {
			log.Fatal(err)
		}
	}
	// Challenges (Commute Green Week is the demo path)
	challenges := []struct {
		id, title, cat, desc, xp, diff, status string
	}{
		{"90000000-0000-7000-8000-000000000001", "Sustainability Sprint", categories[2][0], "Log 5 sustainable actions this week.", "200", "hard", "active"},
		{"90000000-0000-7000-8000-000000000002", "Commute Green Week", categories[2][0], "Cycle or carpool to work for 5 days.", "120", "medium", "active"},
		{"90000000-0000-7000-8000-000000000003", "Recycle Challenge", categories[3][0], "Sort and recycle office waste for a week.", "80", "easy", "under_review"},
	}
	for _, ch := range challenges {
		_, err = tx.Exec(ctx, `INSERT INTO challenges(id,title,category_id,description,xp,difficulty,evidence_required,deadline,status)
			VALUES($1,$2,$3,$4,$5::int,$6,true,'2026-07-25',$7)
			ON CONFLICT(id) DO UPDATE SET title=EXCLUDED.title,status=EXCLUDED.status,xp=EXCLUDED.xp`,
			ch.id, ch.title, ch.cat, ch.desc, ch.xp, ch.diff, ch.status)
		if err != nil {
			log.Fatal(err)
		}
	}
	// Rename Green Starter badge to Green Beginner for demo narrative if present
	_, _ = tx.Exec(ctx, `UPDATE badges SET name='Green Beginner', description='Unlock: 100 XP' WHERE id=$1`, "60000000-0000-7000-8000-000000000001")
	// Trainings
	trainings := [][3]string{
		{"a0000000-0000-7000-8000-000000000001", "ESG Fundamentals", "All employees"},
		{"a0000000-0000-7000-8000-000000000002", "Anti-Corruption Awareness", "All employees"},
		{"a0000000-0000-7000-8000-000000000003", "Carbon Accounting Basics", "Dept heads"},
	}
	for _, t := range trainings {
		_, err = tx.Exec(ctx, `INSERT INTO trainings(id,name,assigned_to,status) VALUES($1,$2,$3,'active') ON CONFLICT(id) DO UPDATE SET name=EXCLUDED.name`, t[0], t[1], t[2])
		if err != nil {
			log.Fatal(err)
		}
	}
	// Give demo employees some points for redeem testing
	_, _ = tx.Exec(ctx, `UPDATE users SET points=1200, xp=50 WHERE role='employee'`)

	if err = tx.Commit(ctx); err != nil {
		log.Fatal(err)
	}
	log.Printf("seed complete: %d departments, %d users, %d factors, phase-3 CSR/challenges", len(departments), len(users), len(factors))
}
func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
