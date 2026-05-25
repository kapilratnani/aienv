package assets

import "github.com/kapilratnani/aienv/internal/registry"

var CuratedMCPServers = []registry.MCPServerItem{
	{Name: "GitHub", Type: "local", Command: "npx -y @modelcontextprotocol/server-github", Description: "Full GitHub API: repos, issues, PRs, Actions"},
	{Name: "PostgreSQL", Type: "local", Command: "npx -y @modelcontextprotocol/server-postgres postgresql://localhost/mydb", Description: "Read-only database access with schema inspection"},
	{Name: "Filesystem", Type: "local", Command: "npx -y @modelcontextprotocol/server-filesystem", Description: "Secure file operations with configurable access"},
	{Name: "Sentry", Type: "local", Command: "npx -y @modelcontextprotocol/server-sentry", Description: "Error tracking, stack traces, frequency trends"},
	{Name: "Docker", Type: "local", Command: "npx -y @docker/mcp-server", Description: "Container management: images, volumes, networks"},
	{Name: "Brave Search", Type: "local", Command: "npx -y @brave/mcp-server", Description: "Web and local search using Brave Search API"},
	{Name: "Playwright", Type: "local", Command: "npx -y @executeautomation/mcp-playwright", Description: "Browser automation: Chrome, Firefox, Safari"},
	{Name: "Kubernetes", Type: "local", Command: "npx -y @kubernetes/mcp-server", Description: "K8s cluster ops: deployments, pods, services, logs"},
	{Name: "Slack", Type: "local", Command: "npx -y @slack/mcp-server", Description: "Channels, messages, search, thread context"},
	{Name: "Stripe", Type: "local", Command: "npx -y @stripe/mcp-server", Description: "Payments, customers, subscriptions, invoices, refunds"},
	{Name: "Linear", Type: "local", Command: "npx -y @linear/mcp-server", Description: "Issue tracking, cycles, projects, sprint management"},
	{Name: "Notion", Type: "local", Command: "npx -y @notion/mcp-server", Description: "Read/write pages, databases, blocks, search"},
	{Name: "Figma", Type: "remote", URL: "https://mcp.figma.com/v1", Command: "https://mcp.figma.com/v1", Description: "File access, component inspection, design tokens"},
	{Name: "Supabase", Type: "local", Command: "npx -y @supabase/mcp-server", Description: "Managed Postgres + auth + storage + edge functions"},
	{Name: "Context7", Type: "local", Command: "npx -y @context7/mcp-server", Description: "Real-time library documentation retrieval"},
	{Name: "PagerDuty", Type: "local", Command: "npx -y @pagerduty/mcp-server", Description: "Incident management, on-call schedules, alerts"},
	{Name: "Datadog", Type: "local", Command: "npx -y @datadog/mcp-server", Description: "Metrics, logs, alerts, APM dashboards"},
	{Name: "Gmail", Type: "local", Command: "npx -y @gmail/mcp-server", Description: "Search, read, send, reply, manage labels"},
	{Name: "Jira", Type: "local", Command: "npx -y @jira/mcp-server", Description: "Issues, workflows, sprints, boards"},
	{Name: "Terraform", Type: "local", Command: "npx -y @terraform/mcp-server", Description: "IaC: plan/apply configurations, manage state"},
}

var CuratedSkills = []registry.SkillItem{
	{Name: "frontend-design", Package: "anthropics/skills", Description: "UI design and frontend implementation", Installs: 421700},
	{Name: "vercel-react-best-practices", Package: "vercel-labs/agent-skills", Description: "React and Next.js best practices", Installs: 389200},
	{Name: "web-design-guidelines", Package: "vercel-labs/agent-skills", Description: "Web design patterns and guidelines", Installs: 317700},
	{Name: "shadcn", Package: "shadcn/ui", Description: "shadcn/ui component patterns", Installs: 147400},
	{Name: "tdd", Package: "mattpocock/skills", Description: "Test-driven development workflow", Installs: 120700},
	{Name: "seo-audit", Package: "coreyhaines31/marketingskills", Description: "SEO audit and optimization", Installs: 112600},
	{Name: "supabase-postgres-best-practices", Package: "supabase/agent-skills", Description: "Postgres and Supabase best practices", Installs: 171800},
	{Name: "better-auth-best-practices", Package: "better-auth/skills", Description: "Authentication best practices", Installs: 51200},
	{Name: "grill-me", Package: "mattpocock/skills", Description: "Code review and design critique", Installs: 171900},
	{Name: "skill-creator", Package: "anthropics/skills", Description: "Create and publish agent skills", Installs: 198100},
	{Name: "pdf", Package: "anthropics/skills", Description: "PDF document generation", Installs: 109700},
	{Name: "pptx", Package: "anthropics/skills", Description: "PowerPoint presentation generation", Installs: 108700},
	{Name: "remotion-best-practices", Package: "remotion-dev/skills", Description: "Remotion video framework patterns", Installs: 299000},
	{Name: "systematic-debugging", Package: "obra/superpowers", Description: "Systematic debugging methodology", Installs: 90000},
	{Name: "to-prd", Package: "mattpocock/skills", Description: "Turn ideas into PRDs", Installs: 80000},
	{Name: "copywriting", Package: "coreyhaines31/marketingskills", Description: "Marketing copywriting", Installs: 60000},
	{Name: "firebase-basics", Package: "firebase/agent-skills", Description: "Firebase fundamentals", Installs: 56300},
	{Name: "azure-ai", Package: "microsoft/azure-skills", Description: "Azure AI services", Installs: 324500},
	{Name: "github-actions-docs", Package: "xixu-me/skills", Description: "GitHub Actions workflow docs", Installs: 45000},
	{Name: "deploy-to-vercel", Package: "vercel-labs/agent-skills", Description: "Deploy to Vercel", Installs: 40000},
}
