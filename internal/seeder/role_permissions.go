package seeder

import (
	"context"
	"fmt"
	"log"
	"strings"
)

// SeedRolePermissions intelligently updates role-permission assignments
// Uses prefix-based assignment: admin.* for SuperAdmin, tenant.* for OrgsAdmin
func (sm *SeedManager) SeedRolePermissions(organizationID int64) error {
	// Get all permissions
	permissions := sm.GetAllPermissions()

	// Define role-permission mappings based on permission prefixes
	rolePermissionMap := map[string][]string{
		"SuperAdmin": {}, // Will include admin.* permissions
		"OrgsAdmin":  {}, // Will include tenant.* permissions
	}

	// Assign permissions based on prefix (scope)
	for _, perm := range permissions {
		switch {
		case strings.HasPrefix(perm.Name, "admin."):
			// SuperAdmin gets all admin.* permissions
			rolePermissionMap["SuperAdmin"] = append(rolePermissionMap["SuperAdmin"], perm.Name)

		case strings.HasPrefix(perm.Name, "tenant."):
			// OrgsAdmin gets all tenant.* permissions
			rolePermissionMap["OrgsAdmin"] = append(rolePermissionMap["OrgsAdmin"], perm.Name)
		}
	}

	// Update role-permission assignments for each role
	for roleName, permissionNames := range rolePermissionMap {
		added, removed, err := sm.updateRolePermissions(roleName, permissionNames, organizationID)
		if err != nil {
			return fmt.Errorf("failed to update permissions for role %s: %v", roleName, err)
		}

		if added > 0 || removed > 0 {
			log.Printf("  Updated %s: +%d permissions, -%d permissions (total: %d)", roleName, added, removed, len(permissionNames))
		} else {
			log.Printf("  %s permissions up to date (%d permissions)", roleName, len(permissionNames))
		}
	}

	log.Println("Successfully updated all role-permission mappings")
	return nil
}

// assignPermissionsToRole assigns specific permissions to a role
func (sm *SeedManager) assignPermissionsToRole(roleName string, permissionNames []string, organizationID int64) error {
	if len(permissionNames) == 0 {
		return nil
	}

	// Get role ID
	var roleID int64
	roleQuery := `SELECT id FROM roles WHERE name = $1 AND organization_id = $2 LIMIT 1`
	err := sm.db.QueryRow(context.Background(), roleQuery, roleName, organizationID).Scan(&roleID)
	if err != nil {
		return fmt.Errorf("failed to get role ID for %s: %v", roleName, err)
	}

	// First, clear existing permissions for this role
	clearQuery := `DELETE FROM role_permissions WHERE role_id = $1`
	_, err = sm.db.Exec(context.Background(), clearQuery, roleID)
	if err != nil {
		return fmt.Errorf("failed to clear existing permissions for role %s: %v", roleName, err)
	}

	// Create placeholders for bulk insert
	valueStrings := make([]string, 0, len(permissionNames))
	valueArgs := make([]interface{}, 0, len(permissionNames)*2)

	for i, permName := range permissionNames {
		valueStrings = append(valueStrings, fmt.Sprintf("($%d, (SELECT id FROM permissions WHERE name = $%d), NOW())", i*2+1, i*2+2))
		valueArgs = append(valueArgs, roleID, permName)
	}

	// Bulk insert role-permission assignments
	insertQuery := fmt.Sprintf(`
		INSERT INTO role_permissions (role_id, permission_id, created_at)
		VALUES %s
		ON CONFLICT (role_id, permission_id) DO NOTHING
	`, strings.Join(valueStrings, ","))

	_, err = sm.db.Exec(context.Background(), insertQuery, valueArgs...)
	if err != nil {
		return fmt.Errorf("failed to insert role-permission assignments: %v", err)
	}

	return nil
}

// updateRolePermissions intelligently updates role permissions (returns added, removed counts)
func (sm *SeedManager) updateRolePermissions(roleName string, permissionNames []string, organizationID int64) (int, int, error) {
	if len(permissionNames) == 0 {
		return 0, 0, nil
	}

	// Get role ID
	var roleID int64
	roleQuery := `SELECT id FROM roles WHERE name = $1 AND organization_id = $2 LIMIT 1`
	err := sm.db.QueryRow(context.Background(), roleQuery, roleName, organizationID).Scan(&roleID)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get role ID for %s: %v", roleName, err)
	}

	// Get currently assigned permissions for this role
	currentPerms := make(map[string]bool)
	currentQuery := `
		SELECT p.name
		FROM permissions p
		JOIN role_permissions rp ON p.id = rp.permission_id
		WHERE rp.role_id = $1
	`
	rows, err := sm.db.Query(context.Background(), currentQuery, roleID)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get current permissions for role %s: %v", roleName, err)
	}
	defer rows.Close()

	for rows.Next() {
		var permName string
		if err := rows.Scan(&permName); err != nil {
			continue
		}
		currentPerms[permName] = true
	}

	// Create sets for new permissions
	newPerms := make(map[string]bool)
	for _, permName := range permissionNames {
		newPerms[permName] = true
	}

	// Find permissions to add (in new but not in current)
	var toAdd []string
	for permName := range newPerms {
		if !currentPerms[permName] {
			toAdd = append(toAdd, permName)
		}
	}

	// Find permissions to remove (in current but not in new)
	var toRemove []string
	for permName := range currentPerms {
		if !newPerms[permName] {
			toRemove = append(toRemove, permName)
		}
	}

	// Add new permissions
	for _, permName := range toAdd {
		addQuery := `
			INSERT INTO role_permissions (role_id, permission_id, created_at)
			VALUES ($1, (SELECT id FROM permissions WHERE name = $2), NOW())
			ON CONFLICT (role_id, permission_id) DO NOTHING
		`
		_, err := sm.db.Exec(context.Background(), addQuery, roleID, permName)
		if err != nil {
			return 0, 0, fmt.Errorf("failed to add permission %s to role %s: %v", permName, roleName, err)
		}
	}

	// Remove obsolete permissions
	for _, permName := range toRemove {
		removeQuery := `
			DELETE FROM role_permissions
			WHERE role_id = $1 AND permission_id = (SELECT id FROM permissions WHERE name = $2)
		`
		_, err := sm.db.Exec(context.Background(), removeQuery, roleID, permName)
		if err != nil {
			return 0, 0, fmt.Errorf("failed to remove permission %s from role %s: %v", permName, roleName, err)
		}
	}

	return len(toAdd), len(toRemove), nil
}

// GetRolePermissionSummary returns a summary of permissions assigned to each role
func (sm *SeedManager) GetRolePermissionSummary(organizationID int64) {
	log.Println("\nROLE-PERMISSION SUMMARY:")
	log.Println(strings.Repeat("=", 50))

	roles := []string{"SuperAdmin", "OrgsAdmin"}

	for _, roleName := range roles {
		query := `
			SELECT p.name
			FROM permissions p
			JOIN role_permissions rp ON p.id = rp.permission_id
			JOIN roles r ON r.id = rp.role_id
			WHERE r.name = $1 AND r.organization_id = $2
			ORDER BY p.name
		`

		rows, err := sm.db.Query(context.Background(), query, roleName, organizationID)
		if err != nil {
			log.Printf("Error fetching permissions for %s: %v", roleName, err)
			continue
		}

		log.Printf("\n%s:", roleName)

		prefixCount := make(map[string]int)
		totalCount := 0

		for rows.Next() {
			var permName string
			err := rows.Scan(&permName)
			if err != nil {
				continue
			}

			// Count by prefix (scope)
			if strings.Contains(permName, ".") {
				prefix := strings.Split(permName, ".")[0]
				prefixCount[prefix]++
			}

			totalCount++
		}
		rows.Close()

		// Show counts by scope
		for prefix, count := range prefixCount {
			visibility := "ALL"
			if prefix == "admin" {
				visibility = "ADMIN_ONLY"
			}
			log.Printf("   %s.* (%s): %d permissions", prefix, visibility, count)
		}

		log.Printf("   Total: %d permissions", totalCount)
	}

	log.Println("\n" + strings.Repeat("=", 50))
}
