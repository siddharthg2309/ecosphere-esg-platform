# EcoSphere: ESG Management Platform

Environmental, Social and Governance (ESG) has become a critical aspect of modern businesses. Organizations are expected to monitor carbon emissions, promote employee well-being, and maintain governance compliance. While many ERP systems collect operational data, ESG reporting is often manual, disconnected, and difficult to monitor in real time.

EcoSphere aims to integrate ESG directly into day-to-day ERP operations by measuring sustainability metrics, encouraging employee participation, and providing meaningful reports for management.

Build an ESG Management Platform that enables organizations to measure, manage and improve their Environmental, Social and Governance performance. The platform should integrate operational data, employee participation and compliance activities into a unified dashboard while encouraging sustainability through gamification.

- Environmental: Carbon accounting, emission factors, sustainability goals and carbon reports
- Social: CSR activities, employee participation, diversity metrics and engagement
- Governance: Policies, audits, compliance tracking and governance reports
- Gamification: Challenges, badges, XP, rewards and leaderboards

## 1. Background

## 2. Challenge Statement

## 3. Core Modules

The application consists of Master Data and Transactional Data.

### Master Data

| Model | Purpose | Key Fields |
| --- | --- | --- |
| Department | Organizational hierarchy and ESG ownership | Name, Code, Head, Parent Department, Employee Count, Status |
| Category | Shared category values used across Social and Gamification modules (e.g., CSR Activity Category, Challenge Category) | Name, Type (CSR Activity / Challenge), Status |
| Emission Factor | Carbon values used during calculations | - |
| Product ESG Profile | ESG information linked to products | - |
| Environmental Goal | Sustainability targets | - |
| ESG Policy | Governance policies | - |
| Badge | Employee achievements | Name, Description, Unlock Rule, Icon |
| Reward | Redeemable incentives | Name, Description, Points Required, Stock, Status |

## 4. Suggested Data Model

### Transactional Data

| Model | Purpose | Key Fields |
| --- | --- | --- |
| Carbon Transaction | Stores calculated emissions from ERP operations | - |
| CSR Activity | Social initiatives organized by the company | - |
| Employee Participation | Tracks employee involvement in CSR Activities only | Employee, Activity, Proof, Approval Status, Points Earned, Completion Date |
| Challenge | Sustainability challenges | Title, Category, Description, XP, Difficulty, Evidence Required, Deadline, Status (Draft / Active / Under Review / Completed / Archived) |
| Challenge Participation | Tracks employee progress within Challenges only | Challenge, Employee, Progress, Proof, Approval, XP Awarded |
| Policy Acknowledgement | Employee policy acceptance | - |
| Audit | Governance audits | - |
| Compliance Issue | Governance violations | Audit, Severity, Description, Owner, Due Date, Status |
| Department Score | Aggregated ESG performance per department | Department, Environmental Score, Social Score, Governance Score, Total Score |

## 5. Business Workflow

```text
Master Configuration
    |
    v
Departments · Categories · Emission Factors · Products
Goals · Policies · Challenges
    |
    v
Daily Business Operations
(Purchase • Manufacturing • Expenses • Fleet)
    |
    v
Carbon Transactions
    |
    v
Employee Participation (CSR) · Challenge Participation
Policy Acknowledgements · Audits
    |
    v
Environmental Score   Social Score   Governance Score
    |
    v
Department Total Score
    |
    v
Overall ESG Score
(weighted average of Department Total Scores - default weighting:
Environmental 40% / Social 30% / Governance 30%, configurable per organization)
    |
    v
Organization Dashboard & Reports
```

## 6. Expected Features

### Environmental

- Configure Emission Factors
- Calculate Carbon Emissions
- Department Carbon Tracking
- Sustainability Goals
- Environmental Dashboard

### Social

- CSR Activities
- Employee Participation
- Diversity Metrics
- Training Completion

### Governance

- ESG Policies
- Policy Acknowledgements
- Audits
- Compliance Issues

### Gamification

- Challenges (with full lifecycle: Draft -> Active -> Under Review -> Completed, or Archived at any point)
- XP
- Badges (auto-awarded when an employee's XP or completed-challenge count satisfies the Badge's Unlock Rule)
- Rewards (redeemable using earned XP/Points - see Section 8)
- Leaderboards

### Settings & Administration

- Departments management
- Category management
- ESG Configuration (see business rules below)
- Notification Settings (see Section 8)

The platform should generate:

- Environmental Report
- Social Report
- Governance Report
- ESG Summary Report
- Custom Report Builder - lets users build a report by combining filters below and export it (PDF / Excel / CSV)

Each report should support filtering by:

- Department
- Date Range
- Module
- Employee
- Challenge
- ESG Category

## 7. Reports

## 8. Core Configuration & Business Rules

The following are in scope, not optional, since they directly support core modules:

- Reward Redemption: Employees can redeem earned Points/XP for a Reward from the catalog, subject to stock availability. Redeeming a Reward deducts the corresponding Points from the employee's balance.
- Notification System: The platform sends notifications (in-app and/or email) for at least: new compliance issue raised, CSR/Challenge approval decisions, policy acknowledgement reminders, and badge unlocks. Configurable via Settings -> Notification Settings.
- Auto Emission Calculation: When enabled (Settings toggle), Carbon Transactions are calculated automatically from linked Purchase/Manufacturing/Expense/Fleet records using the relevant Emission Factor - no manual entry required.
- Evidence Requirement: When enabled (Settings toggle), CSR Activity participation cannot be marked Approved without an attached proof file.
- Badge Auto-Award: When enabled (Settings toggle), a Badge is automatically assigned to an employee the moment their XP, completed-challenge count, or other tracked metric satisfies that Badge's Unlock Rule - no manual admin action required.
- Compliance Issue Ownership: Every Compliance Issue must have an assigned Owner and a Due Date; issues that pass their Due Date while still Open should be flagged (feeds the Notification System above).

Participants are encouraged to explore additional features such as:

- Department ESG rankings
- Smart dashboard visualizations
- Mobile-responsive interface

Mockup: https://link.excalidraw.com/l/65VNwvy7c4X/2m6lz9Ln4

## 9. Bonus Ideas (Optional)
