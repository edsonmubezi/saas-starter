package handler

import (
	"log"
	"net/http"

	"github.com/edsonmubezi/myapp/pkg/database"
)

type MonthlyCount struct {
	Month string `json:"month"`
	Count int    `json:"count"`
}

type AdminDashboardData struct {
	TotalOrganizations  int            `json:"total_organizations"`
	TotalUsers          int            `json:"total_users"`
	ActiveUsers30Days   int            `json:"active_users_30_days"`
	FailedLoginAttempts int            `json:"failed_login_attempts"`
	LockedAccounts      int            `json:"locked_accounts"`
	UserGrowthTrend     []MonthlyCount `json:"user_growth_trend"`
}

func GetAdminDashboardHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	db := database.DB

	var d AdminDashboardData

	// Total organizations
	if err := db.QueryRow(ctx, `SELECT COUNT(*) FROM organizations WHERE deleted_at IS NULL`).Scan(&d.TotalOrganizations); err != nil {
		log.Printf("dashboard: org count error: %v", err)
	}

	// Total users
	if err := db.QueryRow(ctx, `SELECT COUNT(*) FROM users WHERE deleted_at IS NULL`).Scan(&d.TotalUsers); err != nil {
		log.Printf("dashboard: user count error: %v", err)
	}

	// Active users in last 30 days
	if err := db.QueryRow(ctx, `SELECT COUNT(*) FROM users WHERE deleted_at IS NULL AND updated_at >= NOW() - INTERVAL '30 days'`).Scan(&d.ActiveUsers30Days); err != nil {
		log.Printf("dashboard: active user count error: %v", err)
	}

	// Failed login attempts in last 24 hours
	if err := db.QueryRow(ctx, `SELECT COUNT(*) FROM login_attempts WHERE success = false AND created_at >= NOW() - INTERVAL '24 hours'`).Scan(&d.FailedLoginAttempts); err != nil {
		log.Printf("dashboard: failed login count error: %v", err)
	}

	// Locked accounts
	if err := db.QueryRow(ctx, `SELECT COUNT(*) FROM users WHERE locked_at IS NOT NULL AND deleted_at IS NULL`).Scan(&d.LockedAccounts); err != nil {
		log.Printf("dashboard: locked accounts error: %v", err)
	}

	// User growth trend — new users per month for last 12 months
	rows, err := db.Query(ctx, `
		SELECT
			TO_CHAR(DATE_TRUNC('month', created_at), 'YYYY-MM') AS month,
			COUNT(*) AS count
		FROM users
		WHERE created_at >= NOW() - INTERVAL '12 months'
		  AND deleted_at IS NULL
		GROUP BY DATE_TRUNC('month', created_at)
		ORDER BY DATE_TRUNC('month', created_at)
	`)
	if err != nil {
		log.Printf("dashboard: growth trend error: %v", err)
	} else {
		defer rows.Close()
		for rows.Next() {
			var mc MonthlyCount
			if err := rows.Scan(&mc.Month, &mc.Count); err == nil {
				d.UserGrowthTrend = append(d.UserGrowthTrend, mc)
			}
		}
	}

	if d.UserGrowthTrend == nil {
		d.UserGrowthTrend = []MonthlyCount{}
	}

	SendJSONResponse(w, http.StatusOK, "Dashboard data retrieved", d)
}
