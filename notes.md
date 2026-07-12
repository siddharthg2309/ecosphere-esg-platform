# constraints  
- proper DB design
- input validation 
- consistent ui design
- notifs system
        Notification System
            │
            ├── New Compliance Issue Raised
            ├── CSR / Challenge Approval Decision
            ├── Policy Acknowledgement Reminder
            └── Badge Unlock


# portal


four portal / role structure
1. employee
2. operational portal (department head)
3. auditor
4. admin (esg admin + management)


environmental emissions 

-> trucks -> 268 kg co2 emit
to reduce it by 20%. EV trucks 

electricity -> Solar 
manufacturing -> new equipments adoption


#employee portal 
Role based access:

employees in department:
    - challenges. (join challenges -> particpate -> verification system -> profile update)
    - policy acknowlegement when sign up
    - employee sustainablity profile
    - department leaderboard
    - employee uploads photo/proof for verification with timestamp -> human approval (AI-assisted check, not the sole gate). points auto-awarded only after approval
    - challenges completion badge and xps
    - personalized challenges based on departement
    - rewards (reedemable giftcards -> ESG budget).   xps - > reward points for redeeming giftcards
    - document upload from employe portal regarding the operational activities occuring in the department. (example: logistics -> truck fuel added by the employee, electricity(maintanance dept)-> kwh used). AI-assisted categorization only; emission value = quantity x emission factor (deterministic, not AI-guessed)
    - compliance issues (current, completed)
    - csr activity participation (join csr activity -> upload proof -> approval -> points earned)   [separate flow from challenges]
    - training completion (assigned trainings, mark complete)


#department portal (operational portal)

 - add new challenges
 - approval of the operational activities. -> after approval, the emmision score is auto-updated
 - overall department leaderboard (trendsm charts)
 - view employeee performance (for their department), participation
 - assign compliance issue to a employee in the department (view status)
 - view department esg scores
 - upload proof of procurement purchases for the reduction of emissions based on the assinged budget
 - add employeees to the department (emoployee management)
 

#auditor portal
- raise compliance (with department, severity, due date) issue (view status)

                        Auditor Login
                            │
                            ▼
                        Select Department
                            │
                            ▼
                        View Department Data
                            │
                            ├── Operational Records
                            ├── Carbon Transactions
                            ├── Evidence & Invoices
                            ├── CSR Records
                            ├── Policy Acknowledgements
                            └── Previous Compliance Issues
                            │
                            ▼
                        Conduct Audit
                            │
                            ▼
                        Raise Compliance Issue
                            │
                            ▼
                        Assign to the Department 


#admin portal (ecosphere-esg-platform/wireframe.png)

- view the entire organization (Social, environment, Governance)
- overall esg score = weighted average of department total scores (default env 40% / social 30% / gov 30%, configurable per org)
- department's performance (competitive)
- organizational policy acknowlegement and policy updation (notifs and employee sign in popup)
- reduction in Environmental Emissions Co2 (Goals for the particular department) -> using ESG budget for emission goals
- add department (name, code, head, parent department, employee count, status), assign head
- manage category master (type: csr activity | challenge) -- shared across social + gamification modules
- configure emission factors (per category, unit, kgco2 per unit)
- product esg profiles (esg info linked to products)
- auto calculation of emissions after department head verifies the operational activity. calc is deterministic: emission = activity quantity x emission factor. (AI only assists in categorizing the uploaded document, it does NOT compute the number)
- verify the proofs submitted by the department related to spending of esg budget.
- csr activites and participation filters: department, overall, diversity
- diversity dashboard (diversity metrics) + training completion tracking (social module) 

                        Management
                            │
                        Approves ESG Budget
                            │
                            ▼
                        ESG Admin
                            │
                        Allocates Budget
                            │
                            ├──────────────┬──────────────┬─────────────┐
                            ▼              ▼              ▼
                        Fleet Dept      Manufacturing     Facilities
                            │              │              │
                        Buy EV Trucks   Buy Efficient     Install
                                        Machines          Solar Panels
                            │              │              │
                            └──────────────┴──────────────┘
                                            │
                                            ▼
                                Lower Carbon Emissions
                                            │
                                            ▼
                            Environmental Score Improves


- esg report overall, department based. 

            Reports
            │
            ▼
            Select Report Type
            │
            ├── Environmental Report
            ├── Social Report
            ├── Governance Report
            ├── ESG Summary Report
            └── Custom Report Builder
            │
            ▼
            Apply Filters
            │
            ├── Department
            ├── Date Range
            ├── Module
            ├── Employee
            ├── Challenge
            └── ESG Category
            │
            ▼
            Generate Report
            │
            ├── Preview Report
            ├── Export PDF
            ├── Export Excel
            └── Export CSV


- allocation of esg budget for rewards.
- admin can create global challenges for all the departments


# data model

master data
- department: name, code, head, parent department, employee count, status
- category: name, type (csr activity | challenge), status   -- shared across social + gamification
- emission factor: name, category, unit, kgco2 per unit, status
- product esg profile: product, esg attributes, linked emission factors
- environmental goal: name, department, target co2, current co2, deadline, status
- esg policy: title, body, version, effective date
- badge: name, description, unlock rule (xp threshold / completed-challenge count), icon
- reward: name, description, points required, stock, status

transactional data
- carbon transaction: source (purchase | manufacturing | expense | fleet), quantity, emission factor, computed co2, date, department   -- computed co2 = quantity x emission factor
- csr activity: title, category, evidence required, status
- employee participation (csr only): employee, activity, proof, approval status, points earned, completion date
- challenge: title, category, description, xp, difficulty, evidence required, deadline, status (draft | active | under review | completed | archived)
- challenge participation (challenges only): challenge, employee, progress, proof, approval, xp awarded
- policy acknowledgement: employee, policy, acknowledged at
- audit: title, department, auditor, date, findings, status
- compliance issue: audit, severity, description, owner, due date, status
- department score: department, environmental score, social score, governance score, total score

note: employee participation (csr) and challenge participation are TWO separate models -- csr proof/points vs challenge progress/xp


# scoring

- department total score = environmental + social + governance (per department)
- overall esg score = weighted average of department total scores
    default weighting: environmental 40% / social 30% / governance 30%   (configurable per organization)


# settings / esg configuration (admin)

toggles (in scope, not optional -- see wireframe screen 7):
- [toggle] enable auto emission calculation -> carbon transactions auto-generated from purchase / manufacturing / expense / fleet using the emission factor (no manual entry)
- [toggle] require evidence for all csr activities -> csr participation cannot be marked approved without an attached proof file
- [toggle] auto-award badges -> badge auto-assigned the moment xp / completed-challenge count satisfies its unlock rule
- [toggle] email / in-app alerts for new compliance issues
- notification settings (channels for the 4 notif types listed under constraints)
- category management + department management
- configurable esg weightings (env / social / gov)


# business rules

- reward redemption: employee redeems points/xp for a reward from the catalog, subject to stock availability. redeeming deducts points from the balance and decrements stock.
- compliance issue ownership: every issue must have an owner + due date. issues past their due date while still open are flagged -> feeds the notification system.
- challenge lifecycle: draft -> active -> under review -> completed (archived at any point).
- badge auto-award: driven by the unlock rule (xp threshold or completed-challenge count).
- evidence requirement: enforced on csr participation approval when the toggle is on.


# modules (ui is organized by module per wireframe.png; portals above map onto these via role-based access)

- dashboard: environmental / social / governance / overall esg score tiles, emissions trend, department esg ranking, recent activity, quick actions
- environmental: emission factors, product esg profiles, carbon transactions, environmental goals
- social: csr activities, employee participation, diversity dashboard, training completion
- governance: policies, policy acknowledgements, audits, compliance issues
- gamification: challenges, challenge participation, badges, rewards, leaderboard
- reports: environmental / social / governance / esg summary / custom report builder
- settings: departments, categories, esg configuration, notification settings  