import * as React from 'react'

type AttributeIconProps = {
  src: string
  label: string
  size?: 'sm' | 'md'
  showLabel?: boolean
}

export function AttributeIcon({ src, label, size = 'sm', showLabel = false }: AttributeIconProps) {
  const [hasImageError, setHasImageError] = React.useState(false)
  const iconSize = size === 'md' ? 'size-5' : 'size-4'
  const textSize = size === 'md' ? 'text-xs' : 'text-[0.62rem]'

  if (!src || hasImageError) {
    return (
      <span className={`archive-badge inline-flex rounded-full border border-border/65 bg-background/42 px-2.5 py-1 ${textSize} font-medium text-foreground`}>
        {label}
      </span>
    )
  }

  return (
    <span
      className={`archive-badge inline-flex items-center gap-1.5 rounded-full border border-border/65 bg-background/42 px-2 py-1 ${textSize} font-medium text-foreground`}
      title={label}
    >
      <img className={`${iconSize} shrink-0`} src={src} alt="" loading="lazy" decoding="async" onError={() => setHasImageError(true)} />
      {showLabel ? <span>{label}</span> : <span className="sr-only">{label}</span>}
    </span>
  )
}
