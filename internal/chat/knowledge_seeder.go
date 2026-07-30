package chat

import (
	"context"
	"fmt"
	"log"
	"strings"
)

type seedArticle struct {
	Title    string
	Category string
	Content  string
}

// SeedGlobalKnowledge populates the knowledge base with default system-wide articles.
// Articles with organization_id = NULL are global and available to all orgs.
func SeedGlobalKnowledge(ctx context.Context, repo KnowledgeRepository, svc EmbeddingService) error {
	articles := getDefaultKnowledgeArticles()

	var allArticles []KnowledgeArticle
	var allTexts []string

	for _, a := range articles {
		chunks := chunkText(a.Content, 2000, 200)
		for i, chunk := range chunks {
			allArticles = append(allArticles, KnowledgeArticle{
				OrganizationID: nil,
				Category:       a.Category,
				Title:          a.Title,
				Content:        chunk,
				ChunkIndex:     i,
				SourceDocID:    slugify(a.Title),
			})
			allTexts = append(allTexts, chunk)
		}
	}

	log.Printf("Seeding %d knowledge chunks from %d articles...", len(allTexts), len(articles))

	// Batch embed in groups of 50
	batchSize := 50
	var allEmbeddings [][]float32
	for i := 0; i < len(allTexts); i += batchSize {
		end := i + batchSize
		if end > len(allTexts) {
			end = len(allTexts)
		}
		batch := allTexts[i:end]
		embeddings, err := svc.EmbedBatch(ctx, batch)
		if err != nil {
			return fmt.Errorf("embedding batch %d-%d: %w", i, end, err)
		}
		allEmbeddings = append(allEmbeddings, embeddings...)
	}

	if err := repo.BulkInsert(ctx, allArticles, allEmbeddings); err != nil {
		return fmt.Errorf("bulk insert: %w", err)
	}

	log.Printf("Successfully seeded %d knowledge chunks", len(allArticles))
	return nil
}

// chunkText splits text into chunks of maxChars with overlap.
func chunkText(text string, maxChars, overlap int) []string {
	if len(text) <= maxChars {
		return []string{text}
	}

	var chunks []string
	start := 0
	for start < len(text) {
		end := start + maxChars
		if end > len(text) {
			end = len(text)
		}
		// Try to break at sentence boundary
		if end < len(text) {
			lastPeriod := strings.LastIndex(text[start:end], ". ")
			if lastPeriod > maxChars/2 {
				end = start + lastPeriod + 2
			}
		}
		chunks = append(chunks, text[start:end])
		start = end - overlap
		if start >= len(text) {
			break
		}
	}
	return chunks
}

func slugify(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "-")
	s = strings.ReplaceAll(s, "'", "")
	return s
}

func getDefaultKnowledgeArticles() []seedArticle {
	return []seedArticle{
		{
			Title:    "Organization Setup and Structure",
			Category: "module_guide",
			Content: `Setting up your organization structure in the Microfinance System is essential for proper operations and reporting.

## Organization Types
The Microfinance System supports four organization types:
- Single Company: Standard organization with one entity
- Multiple Company: Organization managing multiple company entities
- Multi-Branch: Organization with multiple branches/locations
- Outsourcing: Agency managing operations for client companies

## Departments
Create departments at Organization > Organization Setup > Departments. Departments are the primary organizational unit. For outsourcing orgs, departments can be linked to specific client companies.

## Positions
Define positions within departments at Organization > Organization Setup > Positions. Each position belongs to a department. Positions represent job roles within the organization (e.g., Branch Manager, Loan Officer, Collections Officer).

## Designations
Create designations (job titles) linked to positions at Organization > Organization Setup > Designations. Designations are the most specific level of the organization hierarchy (e.g., Senior Loan Officer, Junior Accountant).

## Document Types
Configure categories for documents (client identification documents, loan agreements, collateral documents, etc.).

## Client Companies (Outsourcing Only)
For outsourcing organizations, manage the client companies you provide services for. Each company can have its own:
- Company-specific settings
- Zones and clusters for geographic organization
- Department and position structure

## Data Import
Bulk import data via Excel templates. The import process includes column mapping, validation, and preview before committing.`,
		},
		{
			Title:    "Communication and Branding Settings",
			Category: "module_guide",
			Content: `Communication features in the Microfinance System help you stay connected with your team and customize system branding.

## Email Broadcasts
Create and send targeted emails at Settings > Communication > Email Broadcasts:
- Compose email with rich text editor
- Filter recipients by department, position, or other criteria
- Schedule or send immediately
- Track delivery status

## Daily Alerts
Set up daily summary emails for administrators. Alerts can include:
- Pending loan applications awaiting approval
- Overdue loan repayments
- Other pending actions requiring attention

## Email Settings
Configure your organization's SMTP email settings at Settings > Communication > Email Settings. Required: SMTP host, port, username, password, sender email address.

## Document Branding
Customize PDF documents generated by the system (loan agreements, reports, statements):
- Header: Organization logo and name
- Footer: Custom text
- Colors: Primary and accent colors
- Watermark: Optional background watermark
- Fonts: Choose from available font families

## Email Branding
Customize email templates:
- Primary and secondary colors
- Font family
- Sign-off text (e.g., "Best regards, Microfinance Team")
- Footer text`,
		},
		{
			Title:    "User Management and Permissions",
			Category: "how_to",
			Content: `Managing users and permissions in the Microfinance System controls who can access what in the system.

## Creating Users
Users can be created manually by administrators at People > Manage Users.

## Roles and Permissions
The Microfinance System uses role-based access control (RBAC):
1. Go to People > Manage Users > Roles
2. Create a role (e.g., "Branch Manager", "Loan Officer", "Admin")
3. Assign permissions to the role
4. Assign the role to users

## Permission Categories
Permissions are organized by module:
- tenant.users.* — User management (view, create, edit, delete)
- tenant.settings.* — Organization settings access
- tenant.reports.* — Various report access

## Two-Factor Authentication
Roles can be configured to require 2FA. Users with such roles must set up and use 2FA (TOTP) for added security.

## Deactivating Users
Go to People > Manage Users, find the user, and toggle their active status. Deactivated users cannot log in but their data is preserved.`,
		},
		{
			Title:    "Frequently Asked Questions",
			Category: "faq",
			Content: `Frequently asked questions about the Microfinance System.

## How do I reset a user's password?
Go to People > Manage Users, find the user, and use the "Reset Password" action. The system will send a password reset link to the user's email.

## How do I deactivate a user?
Go to People > Manage Users, find the user, and toggle their active status. Deactivated users cannot log in but their data is preserved.

## How do I export data to Excel?
Most list pages have an "Export" button that generates an Excel file with the current filtered data.

## How do I configure dark mode?
Click the sun/moon icon in the top-right header bar to toggle between light and dark themes. The preference is saved locally.

## Can multiple people use the system at the same time?
Yes, the Microfinance System is a multi-user web application. Multiple users can work simultaneously with their own accounts and permissions.

## How do I back up my data?
Data is stored in PostgreSQL. Database backups should be configured at the server level. Contact your system administrator for backup policies.

## Is my data secure?
Yes. The Microfinance System uses:
- AES-256-GCM encryption for sensitive data (IDs, API keys)
- JWT tokens with blacklist for session management
- Role-based access control for authorization
- Rate limiting to prevent abuse
- Input sanitization (XSS prevention)
- HTTPS encryption for all communications
- Multi-tenancy isolation (organizations cannot see each other's data)

## How do I change my organization's settings?
Organization settings are accessible from the Settings section in the sidebar. You can configure: organization type, session timeout, and more.

## How do I set up email notifications?
Go to Settings > Communication > Email Settings. Configure your SMTP settings (host, port, username, password, sender address). Then set up Daily Alerts for automated notifications about pending actions.

## What organization types are available?
The system supports four types: Single Company, Multiple Company, Multi-Branch, and Outsourcing. Each type determines which features are available. For example, Client Companies and Zones are only available for outsourcing organizations.`,
		},
	}
}
