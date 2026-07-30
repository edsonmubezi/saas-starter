import React from 'react'
import clsx from 'clsx'

type ColorVariant = 'blue' | 'green' | 'yellow' | 'red' | 'purple' | 'cyan' | 'orange' | 'pink'

interface StatItem {
  label: string
  value: number | string
  icon?: React.ReactNode
  color?: ColorVariant
  change?: {
    value: number
    label?: string
  }
}

interface StatsCardProps {
  stat: StatItem
  className?: string
}

interface StatsGridProps {
  stats: StatItem[]
  columns?: 2 | 3 | 4 | 5
  className?: string
}

const colorMap: Record<ColorVariant, { bg: string; border: string; icon: string; text: string }> = {
  blue: {
    bg: 'from-blue-500/10 to-blue-600/5',
    border: 'border-blue-500/20',
    icon: 'text-blue-400',
    text: 'text-blue-400',
  },
  green: {
    bg: 'from-emerald-500/10 to-emerald-600/5',
    border: 'border-emerald-500/20',
    icon: 'text-emerald-400',
    text: 'text-emerald-400',
  },
  yellow: {
    bg: 'from-yellow-500/10 to-yellow-600/5',
    border: 'border-yellow-500/20',
    icon: 'text-yellow-400',
    text: 'text-yellow-400',
  },
  red: {
    bg: 'from-red-500/10 to-red-600/5',
    border: 'border-red-500/20',
    icon: 'text-red-400',
    text: 'text-red-400',
  },
  purple: {
    bg: 'from-purple-500/10 to-purple-600/5',
    border: 'border-purple-500/20',
    icon: 'text-purple-400',
    text: 'text-purple-400',
  },
  cyan: {
    bg: 'from-cyan-500/10 to-cyan-600/5',
    border: 'border-cyan-500/20',
    icon: 'text-cyan-400',
    text: 'text-cyan-400',
  },
  orange: {
    bg: 'from-orange-500/10 to-orange-600/5',
    border: 'border-orange-500/20',
    icon: 'text-orange-400',
    text: 'text-orange-400',
  },
  pink: {
    bg: 'from-pink-500/10 to-pink-600/5',
    border: 'border-pink-500/20',
    icon: 'text-pink-400',
    text: 'text-pink-400',
  },
}

export function StatsCard({ stat, className }: StatsCardProps) {
  const colors = colorMap[stat.color || 'blue']

  return (
    <div
      className={clsx(
        'bg-gradient-to-br border rounded-xl p-5 backdrop-blur',
        colors.bg,
        colors.border,
        className
      )}
    >
      <div className="flex items-center justify-between">
        <div className="flex-1">
          <p className="text-foreground/60 text-sm font-medium">{stat.label}</p>
          <p className="text-3xl font-bold text-foreground mt-1">
            {typeof stat.value === 'number' ? stat.value.toLocaleString() : stat.value}
          </p>
          {stat.change && (
            <p className={clsx('text-xs mt-2', stat.change.value >= 0 ? 'text-emerald-400' : 'text-red-400')}>
              {stat.change.value >= 0 ? '+' : ''}{stat.change.value}%
              {stat.change.label && <span className="text-foreground/40 ml-1">{stat.change.label}</span>}
            </p>
          )}
        </div>
        {stat.icon && (
          <div className={clsx('p-3 rounded-xl bg-foreground/5', colors.icon)}>
            {stat.icon}
          </div>
        )}
      </div>
    </div>
  )
}

export function StatsGrid({ stats, columns = 4, className }: StatsGridProps) {
  const gridCols = {
    2: 'grid-cols-1 sm:grid-cols-2',
    3: 'grid-cols-1 sm:grid-cols-2 lg:grid-cols-3',
    4: 'grid-cols-1 sm:grid-cols-2 lg:grid-cols-4',
    5: 'grid-cols-1 sm:grid-cols-2 lg:grid-cols-5',
  }

  return (
    <div className={clsx('grid gap-4', gridCols[columns], className)}>
      {stats.map((stat, idx) => (
        <StatsCard key={idx} stat={stat} />
      ))}
    </div>
  )
}

export default StatsCard
