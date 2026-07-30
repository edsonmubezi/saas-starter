// /rootfolder/utils/dashboard.ts
export type Stats = {
  totals: { users: number; orgs?: number; activeUsers?: number }
  trend: Array<{ date: string; users: number }>
}

// Fake data generator (12 months: 2024-09 .. 2025-08 for example)
function makeDemoStats(): Stats {
  const now = new Date()
  const months: { date: string; users: number }[] = []
  for (let i = 11; i >= 0; i--) {
    const d = new Date(now.getFullYear(), now.getMonth() - i, 1)
    const iso = `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}`
    months.push({ date: iso, users: 200 + Math.round(Math.random() * 300) })
  }
  return {
    totals: { users: 1280, orgs: 12, activeUsers: 436 },
    trend: months,
  }
}

// FAKE: return demo data (simulate latency)
export async function getDashboardStats(): Promise<Stats> {
  await new Promise(r => setTimeout(r, 400))
  return makeDemoStats()
}
