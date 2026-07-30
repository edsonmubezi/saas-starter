
import React from 'react'
import AsyncDropdown from '../ui/AsyncDropdown'
import { getOrganizationsdropodown } from '../utils/orgs'

type Org = { id: number; name: string }

export default function OrganizationFilter({
  value,
  onChange,
}: {
  value?: string
  onChange: (v: string) => void
}) {
  return (
    <AsyncDropdown<Org>
      label="Organization"
      value={value}
      onChange={onChange}
      loader={getOrganizationsdropodown}
      getValue={o => String(o.id)}
      getLabel={o => o.name}
      placeholder="All organizations"
      selectClassName="select w-full sm:w-[180px]"
    />
  )
}
