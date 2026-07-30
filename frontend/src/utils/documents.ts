// src/utils/documents.ts
import { request, unwrap } from './common'
import { uploadRequest } from './uploadRequest'
import { API_PREFIX } from './apiPrefix'

const BASE = `${API_PREFIX.ORG}/employee-docs`
/** Types aligned to backend (IDs are strings in React) */
export type Doc = {
  id: string
  employee_id: string
  doctype_id: string
  doctype_name?: string
  docurl?: string
  delete_status?: number
  organization_id?: string
}

export type Id = string | number

/** GET /employee-docs/{employeeId} */
export const listDocuments = (employeeId: Id) =>
  unwrap<Doc[]>(request(`${BASE}/${employeeId}`, { method: 'GET' }))

/** POST /employee-docs/{employeeId}  (multipart/form-data) */
export const createDocument = (
  employeeId: Id,
  payload: { doctype_id: string; file: File }
) => {
  const fd = new FormData()

  // Must match your handler: UploadPDFFile(r, "supporting_document")
  fd.append('supporting_document', payload.file)

  // Go expects numbers for these (int64), not strings
  const docData = {
    id: 0,
    employee_id: Number(employeeId),
    doctype_id: Number(payload.doctype_id),
    docurl: '',
    delete_status: 0,
   
  }
  fd.append('doc_data', JSON.stringify(docData))

 

  return unwrap<Doc>(uploadRequest(`${BASE}/${employeeId}`, fd, { method: 'POST' }))
}

/** DELETE /employee-docs/{docId} */
export const softDeleteDocument = (docId: string) =>
  unwrap<Doc>(request(`${BASE}/${docId}`, { method: 'DELETE' }))
