
# SaaS Starter Admin (Vite + TS)

Features:
- Dark Catalyst-style layout (sidebar + topbar)
- Inter font globally
- Token-based auth with refresh
- React Query data layer
- Reusable DataTable (sticky header, skeleton, zebra, sort, page-size)
- Live search (debounce), org filter, status filter + status badges
- Users CRUD with modal forms
- Details modal with Activate/Deactivate + Send password reset

## Run
```bash
npm i
cp .env.example .env
# set VITE_API_BASE_URL to your Go API base, e.g., http://localhost:8080/api
npm run dev
```

## Expected API
- POST /api/v1/login  → { data: { access_token, refresh_token } }
- POST /api/v1/logout
- POST /api/v1/refresh → { access_token, refresh_token? }
- GET  /api/v1/users/me
- GET  /api/v1/admin/users?page=&page_size=&q=&organization_id=&sort_by=&sort_dir=&status=
- POST /api/v1/admin/users/create
- PUT  /api/v1/admin/users/{id}
- DELETE /api/v1/admin/users/{id}
- POST /api/v1/admin/users/{id}/admin-reset-password
- GET  /api/v1/admin/organizations → [{ id, name }]
- GET  /api/v1/admin/roles         → [{ id, name }]
