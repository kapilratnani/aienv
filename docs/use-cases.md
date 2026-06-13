# Use Cases — Historical

The table below was written for the pre-v2 architecture where agents were defined
as a combination of MCPs, skills, and rules. In the current black-box agent
architecture, the user supplies the entire agent configuration — aienv provides
only the sandbox, permissions, and audit layer.

The scenarios remain valid as motivating examples for what aienv enables:
isolated, reproducible environments for different coding tasks.

| # | Scenario | MCPs | Skills | Rules |
|---|---|---|---|---|
| 1 | **Backend API dev** | postgres, GitHub | api-design | AGENTS.md |
| 2 | **Frontend design** | Figma, Playwright | web-design-guidelines | — |
| 3 | **Incident response** | PagerDuty, Datadog, Slack | — | runbook |
| 4 | **Data pipeline** | Snowflake, dbt | airflow | data-docs |
| 5 | **Security audit** | Snyk, Semgrep | security-review | policy |
| 6 | **Documentation site** | Notion | content-style | doc-standards |
| 7 | **Payment integration** | Stripe | test-plan | compliance |
| 8 | **CI/CD debugging** | GitHub Actions, Docker | devops | runbook |
| 9 | **Database migration** | Supabase | schema-review | migration-guide |
| 10 | **Mobile API design** | Apollo GraphQL | api-design | style-guide |
| 11 | **Code review** | GitHub | review-checklist | PR template |
| 12 | **Local-first app** | SQLite | tdd | architecture |
| 13 | **E2E testing** | Playwright | test-plan | coverage |
| 14 | **CLI tool dev** | (none) | tdd | coding-standards |
| 15 | **Multi-repo project** | GitHub | monorepo | workspace |
| 16 | **Self-referential dev** | GitHub | tdd, grill-me, caveman | AGENTS.md + Obsidian notes |
