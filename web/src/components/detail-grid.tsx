export function DetailGrid({ details }: { details: Array<{ label: string; value: string }> }) {
  return (
    <dl className="grid gap-2 sm:grid-cols-2 xl:grid-cols-4">
      {details.map((detail) => (
        <div
          key={detail.label}
          className="rounded-[0.72rem] border border-border/65 bg-background/42 px-4 py-3 shadow-[inset_0_1px_0_rgba(255,255,255,0.64)]"
        >
          <dt className="text-[0.68rem] font-medium text-muted-foreground">{detail.label}</dt>
          <dd className="mt-1 font-display text-base leading-6 text-foreground">{detail.value}</dd>
        </div>
      ))}
    </dl>
  )
}

export function CompactStatList({ title, entries }: { title: string; entries: Array<[string, number]> }) {
  return (
    <section aria-label={title}>
      <h3 className="mb-3 text-[0.68rem] font-semibold uppercase tracking-[0.18em] text-muted-foreground">{title}</h3>
      <dl className="grid gap-2 sm:grid-cols-2 xl:grid-cols-4">
        {entries.map(([label, value]) => (
          <div
            key={label}
            className="flex items-center justify-between gap-3 rounded-[0.72rem] border border-border/65 bg-background/42 px-4 py-3 shadow-[inset_0_1px_0_rgba(255,255,255,0.58)]"
          >
            <dt className="truncate text-xs text-muted-foreground">{label}</dt>
            <dd className="font-display text-base tabular-nums text-foreground">{value.toLocaleString()}</dd>
          </div>
        ))}
      </dl>
    </section>
  )
}
