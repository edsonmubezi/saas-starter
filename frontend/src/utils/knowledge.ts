import { request, QsFrom } from './common'
import { API_PREFIX } from './apiPrefix'

const BASE = `${API_PREFIX.ORG}/chat/knowledge`

export interface KnowledgeArticle {
  id: string
  organization_id?: number
  category: string
  title: string
  content: string
  chunk_index: number
  source_doc_id: string
  created_at: string
  updated_at: string
}

export interface KnowledgeListResponse {
  status: number
  data: KnowledgeArticle[]
  total: number
  page: number
}

export async function getKnowledgeArticles(params: {
  page?: number
  page_size?: number
  category?: string
}): Promise<KnowledgeListResponse> {
  return request(`${BASE}?${QsFrom(params)}`)
}

export async function createKnowledgeArticle(data: {
  title: string
  content: string
  category: string
}): Promise<{ status: number; data: KnowledgeArticle }> {
  return request(`${BASE}`, {
    method: 'POST',
    body: JSON.stringify(data),
  })
}

export async function updateKnowledgeArticle(
  id: string,
  data: { title?: string; content?: string; category?: string },
): Promise<{ status: number; data: KnowledgeArticle }> {
  return request(`${BASE}/${encodeURIComponent(id)}`, {
    method: 'PUT',
    body: JSON.stringify(data),
  })
}

export async function deleteKnowledgeArticle(id: string): Promise<void> {
  return request(`${BASE}/${encodeURIComponent(id)}`, {
    method: 'DELETE',
  })
}

export async function seedKnowledge(): Promise<{ status: number; message: string }> {
  return request(`${BASE}/seed`, {
    method: 'POST',
  })
}

export const KNOWLEDGE_CATEGORIES = [
  { value: 'module_guide', label: 'Module Guide' },
  { value: 'how_to', label: 'How-To Guide' },
  { value: 'troubleshooting', label: 'Troubleshooting' },
  { value: 'faq', label: 'FAQ' },
  { value: 'policy', label: 'Policy' },
] as const
