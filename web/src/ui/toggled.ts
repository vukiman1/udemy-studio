export function toggled(selected: ReadonlySet<string>, item: string): ReadonlySet<string> {
  const next = new Set(selected)
  if (next.has(item)) next.delete(item)
  else next.add(item)
  return next
}
