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
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
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

	departments := []department{
		{"10000000-0000-7000-8000-000000000001", "Manufacturing", "MFG"},
		{"10000000-0000-7000-8000-000000000002", "Logistics", "LOG"},
		{"10000000-0000-7000-8000-000000000003", "Human Resources", "HR"},
		{"10000000-0000-7000-8000-000000000004", "Finance", "FIN"},
		{"10000000-0000-7000-8000-000000000005", "Compliance", "CMP"},
	}
	for _, d := range departments {
		if _, err = tx.Exec(ctx, `INSERT INTO departments(id,name,code,status,employee_count) VALUES($1,$2,$3,'active',0)
			ON CONFLICT(id) DO UPDATE SET name=EXCLUDED.name,code=EXCLUDED.code`, d.id, d.name, d.code); err != nil {
			log.Fatal(err)
		}
	}

	password := env("SEED_ADMIN_PASSWORD", "ChangeMe123!")
	hash, err := platformauth.HashPassword(password)
	if err != nil {
		log.Fatal(err)
	}
	users := []user{
		{"20000000-0000-7000-8000-000000000001", "Aarav Mehta", env("SEED_ADMIN_EMAIL", "admin@ecosphere.local"), "admin", departments[4].id},
		{"20000000-0000-7000-8000-000000000002", "Sneha Nair", "sneha@ecosphere.local", "dept_head", departments[0].id},
		{"20000000-0000-7000-8000-000000000003", "Rohan Iyer", "rohan@ecosphere.local", "dept_head", departments[1].id},
		{"20000000-0000-7000-8000-000000000004", "Priya Shah", "priya@ecosphere.local", "dept_head", departments[2].id},
		{"20000000-0000-7000-8000-000000000005", "Kiran Menon", "kiran@ecosphere.local", "auditor", departments[4].id},
	}
	for i := 6; i <= 20; i++ {
		dept := departments[(i-1)%len(departments)]
		users = append(users, user{
			fmt.Sprintf("20000000-0000-7000-8000-%012d", i),
			fmt.Sprintf("Demo Employee %02d", i),
			fmt.Sprintf("employee%02d@ecosphere.local", i),
			"employee",
			dept.id,
		})
	}
	for _, u := range users {
		if _, err = tx.Exec(ctx, `INSERT INTO users(id,name,email,password_hash,role,department_id,status)
			VALUES($1,$2,$3,$4,$5,$6,'active')
			ON CONFLICT(id) DO UPDATE SET name=EXCLUDED.name,email=EXCLUDED.email,role=EXCLUDED.role,department_id=EXCLUDED.department_id,password_hash=EXCLUDED.password_hash`,
			u.id, u.name, u.email, hash, u.role, u.department); err != nil {
			log.Fatal(err)
		}
	}
	for i := 0; i < 4; i++ {
		_, _ = tx.Exec(ctx, `UPDATE departments SET head_id=$1 WHERE id=$2`, users[i+1].id, departments[i].id)
	}
	// Headcounts for scoring weight narrative
	for i, d := range departments {
		count := 12 + i*3
		_, _ = tx.Exec(ctx, `UPDATE departments SET employee_count=$2 WHERE id=$1`, d.id, count)
	}

	categories := [][4]string{
		{"30000000-0000-7000-8000-000000000001", "Community Service", "csr_activity", "active"},
		{"30000000-0000-7000-8000-000000000002", "Health & Wellness", "csr_activity", "active"},
		{"30000000-0000-7000-8000-000000000003", "Energy Saving", "challenge", "active"},
		{"30000000-0000-7000-8000-000000000004", "Waste Reduction", "challenge", "active"},
	}
	for _, v := range categories {
		_, err = tx.Exec(ctx, `INSERT INTO categories(id,name,type,status) VALUES($1,$2,$3,$4)
			ON CONFLICT(id) DO UPDATE SET name=EXCLUDED.name,type=EXCLUDED.type,status=EXCLUDED.status`, v[0], v[1], v[2], v[3])
		if err != nil {
			log.Fatal(err)
		}
	}

	factors := [][5]string{
		{"40000000-0000-7000-8000-000000000001", "Diesel", categories[2][0], "litre", "2.6800"},
		{"40000000-0000-7000-8000-000000000002", "Petrol", categories[2][0], "litre", "2.3100"},
		{"40000000-0000-7000-8000-000000000003", "Grid electricity", categories[2][0], "kWh", "0.7100"},
		{"40000000-0000-7000-8000-000000000004", "Natural gas", categories[2][0], "kg", "2.7500"},
		{"40000000-0000-7000-8000-000000000005", "Air travel", categories[2][0], "passenger-km", "0.1580"},
		{"40000000-0000-7000-8000-000000000006", "Rail travel", categories[2][0], "passenger-km", "0.0410"},
		{"40000000-0000-7000-8000-000000000007", "Landfill waste", categories[3][0], "kg", "0.5860"},
		{"40000000-0000-7000-8000-000000000008", "Recycled waste", categories[3][0], "kg", "0.0210"},
	}
	for _, v := range factors {
		_, err = tx.Exec(ctx, `INSERT INTO emission_factors(id,name,category_id,unit,kgco2_per_unit,status)
			VALUES($1,$2,$3,$4,$5,'active')
			ON CONFLICT(id) DO UPDATE SET name=EXCLUDED.name,category_id=EXCLUDED.category_id,unit=EXCLUDED.unit,kgco2_per_unit=EXCLUDED.kgco2_per_unit`,
			v[0], v[1], v[2], v[3], v[4])
		if err != nil {
			log.Fatal(err)
		}
	}

	policies := [][4]string{
		{"50000000-0000-7000-8000-000000000001", "Environmental Responsibility", "Employees must minimize waste and report environmental incidents.", "2026-01-01"},
		{"50000000-0000-7000-8000-000000000002", "Supplier Code of Conduct", "Suppliers must meet EcoSphere environmental and social standards.", "2026-01-01"},
		{"50000000-0000-7000-8000-000000000003", "Ethics and Governance", "All employees must follow the company ethics and governance controls.", "2026-01-01"},
	}
	for _, v := range policies {
		_, err = tx.Exec(ctx, `INSERT INTO esg_policies(id,title,body,version,effective_date) VALUES($1,$2,$3,1,$4)
			ON CONFLICT(id) DO UPDATE SET title=EXCLUDED.title,body=EXCLUDED.body`, v[0], v[1], v[2], v[3])
		if err != nil {
			log.Fatal(err)
		}
	}

	badges := [][6]string{
		{"60000000-0000-7000-8000-000000000001", "Green Beginner", "Unlock: 100 XP", "leaf", "xp", "100"},
		{"60000000-0000-7000-8000-000000000002", "Eco Champion", "Earn 500 XP", "award", "xp", "500"},
		{"60000000-0000-7000-8000-000000000003", "Challenge Builder", "Complete 3 challenges", "target", "challenges", "3"},
		{"60000000-0000-7000-8000-000000000004", "Sustainability Leader", "Complete 10 challenges", "trophy", "challenges", "10"},
	}
	for _, v := range badges {
		_, err = tx.Exec(ctx, `INSERT INTO badges(id,name,description,icon,unlock_rule)
			VALUES($1,$2,$3,$4,jsonb_build_object('type',$5::text,'value',$6::int))
			ON CONFLICT(id) DO UPDATE SET name=EXCLUDED.name,description=EXCLUDED.description,icon=EXCLUDED.icon,unlock_rule=EXCLUDED.unlock_rule`,
			v[0], v[1], v[2], v[3], v[4], v[5])
		if err != nil {
			log.Fatal(err)
		}
	}

	rewards := [][5]string{
		{"70000000-0000-7000-8000-000000000001", "Reusable Bottle", "EcoSphere steel bottle", "150", "25"},
		{"70000000-0000-7000-8000-000000000002", "Plant a Tree", "Tree planted in your name", "250", "100"},
		{"70000000-0000-7000-8000-000000000003", "Gift Card", "Sustainable marketplace gift card", "500", "20"},
		{"70000000-0000-7000-8000-000000000004", "Volunteer Day", "One paid volunteer day", "1000", "10"},
	}
	for _, v := range rewards {
		_, err = tx.Exec(ctx, `INSERT INTO rewards(id,name,description,points_required,stock,status)
			VALUES($1,$2,$3,$4,$5,'active')
			ON CONFLICT(id) DO UPDATE SET name=EXCLUDED.name,description=EXCLUDED.description,points_required=EXCLUDED.points_required,stock=EXCLUDED.stock`,
			v[0], v[1], v[2], v[3], v[4])
		if err != nil {
			log.Fatal(err)
		}
	}

	// Demographics for diversity dashboard
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
		{"80000000-0000-7000-8000-000000000005", "Urban Composting Drive", categories[0][0], "Build community compost pits with local NGOs.", "45", "true"},
	}
	for _, v := range csr {
		_, err = tx.Exec(ctx, `INSERT INTO csr_activities(id,title,category_id,description,points,evidence_required,status,activity_date)
			VALUES($1,$2,$3,$4,$5::int,$6::bool,'active','2026-06-15')
			ON CONFLICT(id) DO UPDATE SET title=EXCLUDED.title,description=EXCLUDED.description,points=EXCLUDED.points`,
			v[0], v[1], v[2], v[3], v[4], v[5])
		if err != nil {
			log.Fatal(err)
		}
	}

	// Challenges
	challenges := []struct {
		id, title, cat, desc, xp, diff, status string
	}{
		{"90000000-0000-7000-8000-000000000001", "Sustainability Sprint", categories[2][0], "Log 5 sustainable actions this week.", "200", "hard", "active"},
		{"90000000-0000-7000-8000-000000000002", "Commute Green Week", categories[2][0], "Cycle or carpool to work for 5 days.", "120", "medium", "active"},
		{"90000000-0000-7000-8000-000000000003", "Recycle Challenge", categories[3][0], "Sort and recycle office waste for a week.", "80", "easy", "under_review"},
		{"90000000-0000-7000-8000-000000000004", "Lights Out Friday", categories[2][0], "Power down non-critical equipment after hours.", "60", "easy", "draft"},
		{"90000000-0000-7000-8000-000000000005", "Zero Single-Use Week", categories[3][0], "Avoid single-use plastics at work for 7 days.", "100", "medium", "completed"},
	}
	for _, ch := range challenges {
		_, err = tx.Exec(ctx, `INSERT INTO challenges(id,title,category_id,description,xp,difficulty,evidence_required,deadline,status)
			VALUES($1,$2,$3,$4,$5::int,$6,true,'2026-07-25',$7)
			ON CONFLICT(id) DO UPDATE SET title=EXCLUDED.title,status=EXCLUDED.status,xp=EXCLUDED.xp,description=EXCLUDED.description`,
			ch.id, ch.title, ch.cat, ch.desc, ch.xp, ch.diff, ch.status)
		if err != nil {
			log.Fatal(err)
		}
	}

	// Trainings
	trainings := [][3]string{
		{"a0000000-0000-7000-8000-000000000001", "ESG Fundamentals", "All employees"},
		{"a0000000-0000-7000-8000-000000000002", "Anti-Corruption Awareness", "All employees"},
		{"a0000000-0000-7000-8000-000000000003", "Carbon Accounting Basics", "Dept heads"},
	}
	for _, t := range trainings {
		_, err = tx.Exec(ctx, `INSERT INTO trainings(id,name,assigned_to,status) VALUES($1,$2,$3,'active')
			ON CONFLICT(id) DO UPDATE SET name=EXCLUDED.name`, t[0], t[1], t[2])
		if err != nil {
			log.Fatal(err)
		}
	}

	// ── Prototype mock data (presentation ready) ───────────────────────────
	// Carbon transactions: verified + draft across departments
	type carbonSeed struct {
		id, dept, source, qty, factorID, factorVal, co2, date, status, verifier string
	}
	diesel := factors[0]
	elec := factors[2]
	landfill := factors[6]
	air := factors[4]
	carbonRows := []carbonSeed{
		{"c0100000-0000-7000-8000-000000000001", departments[1].id, "fleet", "120", diesel[0], diesel[4], "321.600", "2026-05-10", "verified", users[2].id},
		{"c0100000-0000-7000-8000-000000000002", departments[1].id, "fleet", "95", diesel[0], diesel[4], "254.600", "2026-06-02", "verified", users[2].id},
		{"c0100000-0000-7000-8000-000000000003", departments[0].id, "manufacturing", "480", elec[0], elec[4], "340.800", "2026-05-18", "verified", users[1].id},
		{"c0100000-0000-7000-8000-000000000004", departments[0].id, "manufacturing", "510", elec[0], elec[4], "362.100", "2026-06-12", "verified", users[1].id},
		{"c0100000-0000-7000-8000-000000000005", departments[3].id, "expense", "2100", air[0], air[4], "331.800", "2026-04-22", "verified", users[0].id},
		{"c0100000-0000-7000-8000-000000000006", departments[2].id, "expense", "18", landfill[0], landfill[4], "10.548", "2026-06-01", "verified", users[3].id},
		{"c0100000-0000-7000-8000-000000000007", departments[4].id, "purchase", "320", elec[0], elec[4], "227.200", "2026-05-28", "verified", users[0].id},
		{"c0100000-0000-7000-8000-000000000008", departments[1].id, "fleet", "40", diesel[0], diesel[4], "0", "2026-07-05", "draft", ""},
		{"c0100000-0000-7000-8000-000000000009", departments[0].id, "manufacturing", "90", elec[0], elec[4], "0", "2026-07-08", "draft", ""},
		{"c0100000-0000-7000-8000-00000000000a", departments[3].id, "expense", "6", landfill[0], landfill[4], "0", "2026-07-10", "draft", ""},
		{"c0100000-0000-7000-8000-00000000000b", departments[1].id, "fleet", "75", diesel[0], diesel[4], "201.000", "2026-03-15", "verified", users[2].id},
		{"c0100000-0000-7000-8000-00000000000c", departments[0].id, "manufacturing", "260", elec[0], elec[4], "184.600", "2026-03-20", "verified", users[1].id},
	}
	for _, r := range carbonRows {
		if r.status == "verified" {
			_, err = tx.Exec(ctx, `INSERT INTO carbon_transactions(
				id,department_id,source,quantity,emission_factor_id,factor_value,computed_co2,txn_date,evidence_url,status,verified_by,verified_at)
				VALUES($1,$2,$3,$4::numeric,$5,$6::numeric,$7::numeric,$8,'prototype://evidence/invoice.pdf','verified',$9,now())
				ON CONFLICT(id) DO UPDATE SET quantity=EXCLUDED.quantity,computed_co2=EXCLUDED.computed_co2,status=EXCLUDED.status,
					verified_by=EXCLUDED.verified_by,verified_at=EXCLUDED.verified_at`,
				r.id, r.dept, r.source, r.qty, r.factorID, r.factorVal, r.co2, r.date, r.verifier)
		} else {
			_, err = tx.Exec(ctx, `INSERT INTO carbon_transactions(
				id,department_id,source,quantity,emission_factor_id,factor_value,computed_co2,txn_date,evidence_url,status)
				VALUES($1,$2,$3,$4::numeric,$5,$6::numeric,0,$7,'prototype://evidence/pending.pdf','draft')
				ON CONFLICT(id) DO UPDATE SET quantity=EXCLUDED.quantity,status='draft',verified_by=NULL,verified_at=NULL`,
				r.id, r.dept, r.source, r.qty, r.factorID, r.factorVal, r.date)
		}
		if err != nil {
			log.Printf("carbon seed: %v", err)
		}
	}

	// Environmental goals
	goals := []struct {
		id, name, dept, target, current, deadline, status string
	}{
		{"d0100000-0000-7000-8000-000000000001", "Reduce Fleet Emissions 20%", departments[1].id, "400.000", "312.000", "2026-12-31", "on_track"},
		{"d0100000-0000-7000-8000-000000000002", "Warehouse Solar Retrofit", departments[0].id, "500.000", "160.000", "2026-09-30", "at_risk"},
		{"d0100000-0000-7000-8000-000000000003", "Paperless Dispatch", departments[1].id, "50.000", "50.000", "2026-06-30", "completed"},
		{"d0100000-0000-7000-8000-000000000004", "Cut Office Energy 15%", departments[4].id, "200.000", "110.000", "2026-11-30", "on_track"},
		{"d0100000-0000-7000-8000-000000000005", "Travel CO₂ Cap", departments[3].id, "300.000", "260.000", "2026-12-31", "at_risk"},
		{"d0100000-0000-7000-8000-000000000006", "HR Hybrid Day Savings", departments[2].id, "80.000", "48.000", "2026-10-31", "on_track"},
	}
	for _, g := range goals {
		_, err = tx.Exec(ctx, `INSERT INTO environmental_goals(id,name,department_id,target_co2,current_co2,deadline,status)
			VALUES($1,$2,$3,$4::numeric,$5::numeric,$6,$7)
			ON CONFLICT(id) DO UPDATE SET name=EXCLUDED.name,target_co2=EXCLUDED.target_co2,current_co2=EXCLUDED.current_co2,status=EXCLUDED.status`,
			g.id, g.name, g.dept, g.target, g.current, g.deadline, g.status)
		if err != nil {
			log.Printf("goal seed: %v", err)
		}
	}

	// CSR participations
	type part struct{ id, emp, act, proof, approval string; points int }
	csrParts := []part{
		{"e0100000-0000-7000-8000-000000000001", users[5].id, csr[0][0], "prototype://proof/tree-emp06.jpg", "approved", 50},
		{"e0100000-0000-7000-8000-000000000002", users[6].id, csr[0][0], "prototype://proof/tree-emp07.jpg", "approved", 50},
		{"e0100000-0000-7000-8000-000000000003", users[7].id, csr[1][0], "prototype://proof/blood-emp08.pdf", "approved", 40},
		{"e0100000-0000-7000-8000-000000000004", users[8].id, csr[2][0], "prototype://proof/beach-emp09.jpg", "pending", 0},
		{"e0100000-0000-7000-8000-000000000005", users[9].id, csr[2][0], "prototype://proof/beach-emp10.jpg", "pending", 0},
		{"e0100000-0000-7000-8000-000000000006", users[10].id, csr[3][0], "", "approved", 30},
		{"e0100000-0000-7000-8000-000000000007", users[11].id, csr[0][0], "prototype://proof/tree-emp12.jpg", "rejected", 0},
		{"e0100000-0000-7000-8000-000000000008", users[5].id, csr[3][0], "", "approved", 30},
		{"e0100000-0000-7000-8000-000000000009", users[12].id, csr[1][0], "prototype://proof/blood-emp13.pdf", "pending", 0},
		{"e0100000-0000-7000-8000-00000000000a", users[13].id, csr[4][0], "prototype://proof/compost-emp14.jpg", "approved", 45},
	}
	for _, p := range csrParts {
		_, err = tx.Exec(ctx, `INSERT INTO employee_participations(id,employee_id,activity_id,proof_url,notes,approval,points_earned,completion_date)
			VALUES($1,$2,$3,$4,'Seed participation',$5,$6,'2026-06-20')
			ON CONFLICT(employee_id,activity_id) DO UPDATE SET approval=EXCLUDED.approval,points_earned=EXCLUDED.points_earned,proof_url=EXCLUDED.proof_url`,
			p.id, p.emp, p.act, p.proof, p.approval, p.points)
		if err != nil {
			log.Printf("csr part seed: %v", err)
		}
	}

	// Challenge participations
	type cpart struct {
		id, ch, emp, proof, approval string
		progress, xp                 int
	}
	chalParts := []cpart{
		{"f0100000-0000-7000-8000-000000000001", challenges[0].id, users[5].id, "prototype://proof/sprint-06.pdf", "in_progress", 60, 0},
		{"f0100000-0000-7000-8000-000000000002", challenges[0].id, users[6].id, "prototype://proof/sprint-07.pdf", "pending", 100, 0},
		{"f0100000-0000-7000-8000-000000000003", challenges[1].id, users[5].id, "prototype://proof/commute-06.pdf", "approved", 100, 120},
		{"f0100000-0000-7000-8000-000000000004", challenges[1].id, users[7].id, "prototype://proof/commute-08.pdf", "in_progress", 80, 0},
		{"f0100000-0000-7000-8000-000000000005", challenges[1].id, users[8].id, "prototype://proof/commute-09.pdf", "pending", 100, 0},
		{"f0100000-0000-7000-8000-000000000006", challenges[2].id, users[9].id, "prototype://proof/recycle-10.pdf", "pending", 100, 0},
		{"f0100000-0000-7000-8000-000000000007", challenges[4].id, users[5].id, "prototype://proof/zerouse-06.pdf", "approved", 100, 100},
		{"f0100000-0000-7000-8000-000000000008", challenges[4].id, users[10].id, "prototype://proof/zerouse-11.pdf", "approved", 100, 100},
		{"f0100000-0000-7000-8000-000000000009", challenges[0].id, users[11].id, "prototype://proof/sprint-12.pdf", "approved", 100, 200},
		{"f0100000-0000-7000-8000-00000000000a", challenges[1].id, users[12].id, "prototype://proof/commute-13.pdf", "rejected", 40, 0},
	}
	for _, p := range chalParts {
		_, err = tx.Exec(ctx, `INSERT INTO challenge_participations(id,challenge_id,employee_id,progress,proof_url,approval,xp_awarded)
			VALUES($1,$2,$3,$4,$5,$6,$7)
			ON CONFLICT(challenge_id,employee_id) DO UPDATE SET progress=EXCLUDED.progress,approval=EXCLUDED.approval,xp_awarded=EXCLUDED.xp_awarded,proof_url=EXCLUDED.proof_url`,
			p.id, p.ch, p.emp, p.progress, p.proof, p.approval, p.xp)
		if err != nil {
			log.Printf("challenge part seed: %v", err)
		}
	}

	// Employee badges
	badgeAwards := [][3]string{
		{"aa100000-0000-7000-8000-000000000001", users[5].id, badges[0][0]},
		{"aa100000-0000-7000-8000-000000000002", users[5].id, badges[2][0]},
		{"aa100000-0000-7000-8000-000000000003", users[6].id, badges[0][0]},
		{"aa100000-0000-7000-8000-000000000004", users[11].id, badges[0][0]},
		{"aa100000-0000-7000-8000-000000000005", users[11].id, badges[1][0]},
		{"aa100000-0000-7000-8000-000000000006", users[10].id, badges[0][0]},
	}
	for _, b := range badgeAwards {
		_, err = tx.Exec(ctx, `INSERT INTO employee_badges(id,employee_id,badge_id) VALUES($1,$2,$3)
			ON CONFLICT(employee_id,badge_id) DO NOTHING`, b[0], b[1], b[2])
		if err != nil {
			log.Printf("badge seed: %v", err)
		}
	}

	// Reward redemptions
	_, _ = tx.Exec(ctx, `INSERT INTO reward_redemptions(id,employee_id,reward_id,points_spent)
		VALUES
		('ab100000-0000-7000-8000-000000000001',$1,$2,150),
		('ab100000-0000-7000-8000-000000000002',$3,$4,250)
		ON CONFLICT(id) DO NOTHING`,
		users[5].id, rewards[0][0], users[11].id, rewards[1][0])
	_, _ = tx.Exec(ctx, `UPDATE rewards SET stock = GREATEST(stock-1,0) WHERE id IN ($1,$2)`, rewards[0][0], rewards[1][0])

	// Training completions (for diversity / training % dashboards)
	// Complete ESG Fundamentals for most employees
	for i := 6; i <= 18; i++ {
		empID := fmt.Sprintf("20000000-0000-7000-8000-%012d", i)
		_, _ = tx.Exec(ctx, `INSERT INTO training_completions(id,employee_id,training_id,completed_at)
			VALUES($1,$2,$3,'2026-05-15') ON CONFLICT(employee_id,training_id) DO NOTHING`,
			fmt.Sprintf("ac100000-0000-7000-8000-%012d", i), empID, trainings[0][0])
	}
	for i := 6; i <= 14; i++ {
		empID := fmt.Sprintf("20000000-0000-7000-8000-%012d", i)
		_, _ = tx.Exec(ctx, `INSERT INTO training_completions(id,employee_id,training_id,completed_at)
			VALUES($1,$2,$3,'2026-06-01') ON CONFLICT(employee_id,training_id) DO NOTHING`,
			fmt.Sprintf("ad100000-0000-7000-8000-%012d", i), empID, trainings[1][0])
	}

	// Rich XP / points for leaderboard
	xpPoints := map[string][2]int{
		users[5].id:  {420, 980},  // emp06
		users[6].id:  {180, 700},  // emp07
		users[7].id:  {240, 640},  // emp08
		users[8].id:  {90, 400},   // emp09
		users[9].id:  {150, 520},  // emp10
		users[10].id: {310, 800},  // emp11
		users[11].id: {610, 1100}, // emp12 leaderboard top
		users[12].id: {70, 300},
		users[13].id: {200, 550},
	}
	for id, v := range xpPoints {
		_, _ = tx.Exec(ctx, `UPDATE users SET xp=$2, points=$3, completed_challenges=GREATEST(completed_challenges,1) WHERE id=$1`, id, v[0], v[1])
	}
	_, _ = tx.Exec(ctx, `UPDATE users SET points=1200, xp=50 WHERE role='employee' AND xp < 50`)
	// Named heroes
	_, _ = tx.Exec(ctx, `UPDATE users SET name='Aditi Rao', xp=3910, points=1400, completed_challenges=4 WHERE id=$1`, users[5].id)
	_, _ = tx.Exec(ctx, `UPDATE users SET xp=5200, points=1600, completed_challenges=6 WHERE id=$1`, users[11].id)

	// Policy acknowledgements — most employees acked, leave a few pending for modal demo
	// Leave employee06 (users[5]) without Ethics policy so policy sign-off has something, or ack all for smoother demo
	// For presentation: ack all policies for admin/auditor/dept; employees ack most
	for i, u := range users {
		for pi, pol := range policies {
			// skip one policy for employee07 so employee portal can show pending if needed
			if u.id == users[6].id && pi == 2 {
				continue
			}
			ackID := fmt.Sprintf("ae100000-0000-7000-8000-%012d", i*10+pi+1)
			_, _ = tx.Exec(ctx, `INSERT INTO policy_acknowledgements(id,employee_id,policy_id,version,acknowledged_at)
				VALUES($1,$2,$3,1,now() - interval '2 days')
				ON CONFLICT(employee_id,policy_id,version) DO NOTHING`,
				ackID, u.id, pol[0])
		}
	}

	// Audits + compliance issues
	auditorID := users[4].id
	audits := []struct {
		id, title, dept, date, findings, status string
	}{
		{"b0000000-0000-7000-8000-000000000001", "Q2 Waste Audit", departments[0].id, "2026-06-12", "3 minor compliance issues identified", "completed"},
		{"b0000000-0000-7000-8000-000000000003", "Vendor Compliance Check", departments[3].id, "2026-07-10", "1 open issue — supplier CoC lag", "under_review"},
		{"b0000000-0000-7000-8000-000000000004", "Energy Governance Review", departments[1].id, "2026-07-22", "", "draft"},
		{"b0000000-0000-7000-8000-000000000005", "Data Retention Audit", departments[4].id, "2026-07-28", "2 minor findings", "under_review"},
		{"b0000000-0000-7000-8000-000000000006", "Fleet Safety & ESG Spot Check", departments[1].id, "2026-05-30", "No critical findings", "completed"},
	}
	for _, a := range audits {
		_, err = tx.Exec(ctx, `INSERT INTO audits(id,title,department_id,auditor_id,audit_date,findings,status)
			VALUES($1,$2,$3,$4,$5,$6,$7)
			ON CONFLICT(id) DO UPDATE SET title=EXCLUDED.title,findings=EXCLUDED.findings,status=EXCLUDED.status`,
			a.id, a.title, a.dept, auditorID, a.date, a.findings, a.status)
		if err != nil {
			log.Printf("audit seed: %v", err)
		}
	}

	issues := []struct {
		id, audit, dept, sev, desc, owner, due, status string
	}{
		{"b0000000-0000-7000-8000-000000000002", audits[0].id, departments[0].id, "high", "Missing MSDS sheets for chemical compounds", users[1].id, "2026-07-02", "open"},
		{"b0000000-0000-7000-8000-000000000007", audits[1].id, departments[3].id, "medium", "Vendor Code of Conduct lag for 2 suppliers", users[0].id, "2026-07-20", "in_progress"},
		{"b0000000-0000-7000-8000-000000000008", audits[3].id, departments[4].id, "low", "Retention policy gap for contractor records", users[1].id, "2026-08-01", "open"},
		{"b0000000-0000-7000-8000-000000000009", audits[0].id, departments[0].id, "medium", "Hazardous waste labels incomplete on line B", users[1].id, "2026-07-18", "in_progress"},
		{"b0000000-0000-7000-8000-00000000000a", "", departments[1].id, "high", "Overdue diesel spill kit inspection", users[2].id, "2026-07-05", "open"},
	}
	for _, iss := range issues {
		if iss.audit == "" {
			_, err = tx.Exec(ctx, `INSERT INTO compliance_issues(id,audit_id,department_id,severity,description,owner_id,due_date,status)
				VALUES($1,NULL,$2,$3,$4,$5,$6,$7)
				ON CONFLICT(id) DO UPDATE SET description=EXCLUDED.description,status=EXCLUDED.status,due_date=EXCLUDED.due_date`,
				iss.id, iss.dept, iss.sev, iss.desc, iss.owner, iss.due, iss.status)
		} else {
			_, err = tx.Exec(ctx, `INSERT INTO compliance_issues(id,audit_id,department_id,severity,description,owner_id,due_date,status)
				VALUES($1,$2,$3,$4,$5,$6,$7,$8)
				ON CONFLICT(id) DO UPDATE SET description=EXCLUDED.description,status=EXCLUDED.status,due_date=EXCLUDED.due_date`,
				iss.id, iss.audit, iss.dept, iss.sev, iss.desc, iss.owner, iss.due, iss.status)
		}
		if err != nil {
			log.Printf("issue seed: %v", err)
		}
	}

	// Department scores (current FY period)
	period := "2026-Q2"
	scores := []struct {
		id, dept string
		env, soc, gov, total int
	}{
		{"af100000-0000-7000-8000-000000000001", departments[0].id, 54, 24, 45, 42}, // MFG
		{"af100000-0000-7000-8000-000000000002", departments[1].id, 61, 38, 70, 56}, // LOG
		{"af100000-0000-7000-8000-000000000003", departments[2].id, 58, 72, 80, 69}, // HR
		{"af100000-0000-7000-8000-000000000004", departments[3].id, 52, 40, 78, 55}, // FIN
		{"af100000-0000-7000-8000-000000000005", departments[4].id, 70, 55, 88, 71}, // CMP
	}
	for _, s := range scores {
		_, err = tx.Exec(ctx, `INSERT INTO department_scores(id,department_id,period,environmental,social,governance,total,computed_at)
			VALUES($1,$2,$3,$4,$5,$6,$7,now())
			ON CONFLICT(department_id,period) DO UPDATE SET environmental=EXCLUDED.environmental,social=EXCLUDED.social,
				governance=EXCLUDED.governance,total=EXCLUDED.total,computed_at=now()`,
			s.id, s.dept, period, s.env, s.soc, s.gov, s.total)
		if err != nil {
			log.Printf("score seed: %v", err)
		}
	}

	// Notifications for each portal persona
	type notif struct{ id, user, typ, title string }
	notifs := []notif{
		{"a0100000-0000-7000-8000-000000000001", users[0].id, "compliance_raised", "High severity issue raised in Manufacturing"},
		{"a0100000-0000-7000-8000-000000000002", users[0].id, "approval_decision", "Monthly ESG score recompute completed"},
		{"a0100000-0000-7000-8000-000000000003", users[2].id, "approval_decision", "2 CSR participations awaiting your approval"},
		{"a0100000-0000-7000-8000-000000000004", users[2].id, "compliance_overdue", "Diesel spill kit inspection is overdue"},
		{"a0100000-0000-7000-8000-000000000005", users[1].id, "compliance_raised", "MSDS sheets issue assigned to you"},
		{"a0100000-0000-7000-8000-000000000006", users[4].id, "compliance_overdue", "Vendor CoC issue approaching due date"},
		{"a0100000-0000-7000-8000-000000000007", users[4].id, "approval_decision", "Energy Governance Review scheduled"},
		{"a0100000-0000-7000-8000-000000000008", users[5].id, "badge_unlock", "Badge unlocked: Green Beginner"},
		{"a0100000-0000-7000-8000-000000000009", users[5].id, "approval_decision", "Commute Green Week participation approved (+120 XP)"},
		{"a0100000-0000-7000-8000-00000000000a", users[5].id, "policy_reminder", "Review updated Supplier Code of Conduct"},
		{"a0100000-0000-7000-8000-00000000000b", users[6].id, "policy_reminder", "Ethics and Governance policy awaits acknowledgement"},
		{"a0100000-0000-7000-8000-00000000000c", users[11].id, "badge_unlock", "Badge unlocked: Eco Champion"},
	}
	for _, n := range notifs {
		_, _ = tx.Exec(ctx, `INSERT INTO notifications(id,user_id,type,title,payload,created_at)
			VALUES($1,$2,$3,$4,'{"source":"seed"}'::jsonb,now() - interval '1 hour')
			ON CONFLICT(id) DO UPDATE SET title=EXCLUDED.title`,
			n.id, n.user, n.typ, n.title)
	}

	// Product ESG profiles
	_, _ = tx.Exec(ctx, `INSERT INTO product_esg_profiles(id,product,attributes,emission_factor_id)
		VALUES
		('a1100000-0000-7000-8000-000000000001','EcoBottle Pro','{"material":"recycled steel","recyclable":true}'::jsonb,$1),
		('a1100000-0000-7000-8000-000000000002','GreenPack Carton','{"material":"FSC paper","recyclable":true}'::jsonb,$2)
		ON CONFLICT(id) DO NOTHING`, factors[7][0], factors[6][0])

	if err = tx.Commit(ctx); err != nil {
		log.Fatal(err)
	}
	log.Printf("seed complete: prototype data for all portals — %d depts, %d users, carbon/goals/CSR/challenges/audits/scores loaded",
		len(departments), len(users))
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}


