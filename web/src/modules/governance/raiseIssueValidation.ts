/** Client validation mirrors domain: owner + due date required. */
export function canSubmitIssue(input: {
  description?: string
  departmentId?: string
  ownerId?: string
  dueDate?: string
  severity?: string
}): boolean {
  return Boolean(
    input.description?.trim() &&
      input.departmentId &&
      input.ownerId &&
      input.dueDate &&
      input.severity,
  )
}
