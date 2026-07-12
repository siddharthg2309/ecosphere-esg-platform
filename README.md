# Ecosphere ESG Platform

Ecosphere ESG Platform helps organizations configure sustainability data, record daily business activity, calculate carbon impact, and track employee participation in ESG and CSR initiatives.

The platform is designed to give sustainability teams a single, structured path from setup to reporting. By enforcing a clear order of operations, it keeps master data consistent, ensures every business activity is traceable to a carbon impact, and makes employee engagement measurable and auditable.

## Required Business Workflow

The platform workflow must follow this order:

```mermaid
flowchart TD
    A[Master Configuration]
    B[Departments · Categories · Emission Factors · Products]
    C[Goals · Policies · Challenges]
    D[Daily Business Operations]
    E[Purchase · Manufacturing · Expenses · Fleet]
    F[Carbon Transactions]
    G[Employee Participation]
    H[CSR Participation · Challenge Participation]
    I[Policy Acknowledgements · Audits]

    A --> B
    B --> C
    C --> D
    D --> E
    E --> F
    F --> G
    G --> H
    H --> I
```

### 1. Master Configuration

Set up the organization-wide ESG foundation before operational data is entered.

- Configure departments, product categories, emission factors, and products.
- Define ESG goals, sustainability policies, and employee challenges.
- Keep master data controlled and reviewed so calculations remain consistent.

### 2. Daily Business Operations

Capture business activities that generate carbon impact.

- Record purchase, manufacturing, expense, and fleet activity.
- Map each activity to the correct department, category, product, and emission factor.
- Ensure operational entries are complete before carbon calculations are finalized.

### 3. Carbon Transactions

Convert verified business operations into measurable carbon records.

- Generate carbon transactions from approved operational activity.
- Track emissions by department, category, product, and business source.
- Use the transaction history for dashboards, reporting, and reduction planning.

### 4. Employee Participation

Connect employees to ESG programs after carbon activity is tracked.

- Manage CSR participation and challenge participation.
- Collect policy acknowledgements from employees.
- Run audits to verify data accuracy, policy compliance, and ESG progress.

## Workflow Rule

Master configuration must be completed before daily business operations are recorded. Operations must be converted into carbon transactions before employee participation, acknowledgements, and audits are evaluated.

## Core Workflow Modules

The platform is organized into four interconnected ESG modules that together support the complete sustainability lifecycle of an organization.

### 🌱 Environmental Management

Responsible for monitoring and reducing the organization's environmental impact.

- Configure emission factors
- Record carbon transactions
- Track department-wise emissions
- Monitor sustainability goals
- Generate environmental reports

### 🤝 Social Management

Encourages employee participation in sustainability initiatives and CSR programs.

- Manage CSR activities
- Track employee participation
- Monitor diversity and engagement metrics
- Record training completion
- Reward participation through gamification

### 🛡️ Governance Management

Ensures organizational compliance with ESG policies and governance standards.

- Manage ESG policies
- Track policy acknowledgements
- Conduct governance audits
- Monitor compliance issues
- Generate governance reports

### 🏆 Gamification

Promotes employee engagement by encouraging participation in ESG initiatives.

- Create sustainability challenges
- Award XP and badges
- Redeem rewards
- Display department and employee leaderboards
- Increase long-term participation through incentives