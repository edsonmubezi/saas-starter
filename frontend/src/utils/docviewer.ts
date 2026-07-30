import { toFileUrl } from './url'

/** Convert a storage path to a viewable URL for iframe/doc display */
export async function fetchDocBlobUrl(storagePath: string): Promise<string> {
  const fullUrl = toFileUrl(storagePath)
  if (!fullUrl) throw new Error('Failed to load document')
  return fullUrl
}
