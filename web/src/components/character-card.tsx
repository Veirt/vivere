import * as React from 'react'
import { Link } from '@tanstack/react-router'

import { AttributeIcon } from '#/components/attribute-icon.tsx'
import { Card, CardContent } from '#/components/ui/card.tsx'
import { displayName, formatElement, formatJob, type Character, type Language } from '#/lib/characters.ts'

type CharacterCardProps = {
  character: Character
  language: Language
}

type PortraitProps = {
  character: Character
  className?: string
  src?: string
}

export function CharacterPortrait({ character, className = '', src }: PortraitProps) {
  const imageSrc = src || character.iconPath
  const [hasImageError, setHasImageError] = React.useState(false)

  React.useEffect(() => {
    setHasImageError(false)
  }, [imageSrc])

  if (!imageSrc || hasImageError) {
    return (
      <div className={`${className} flex items-center justify-center bg-secondary px-3 text-center text-[0.62rem] font-semibold uppercase tracking-[0.16em] text-muted-foreground`} role="img" aria-label={`Artwork unavailable for ${character.name || character.nameKey}`}>
        Artwork unavailable
      </div>
    )
  }

  return (
    <img
      className={className}
      src={imageSrc}
      alt={`Portrait of ${character.name || character.nameKey}`}
      loading="lazy"
      decoding="async"
      onError={() => setHasImageError(true)}
    />
  )
}

export function CharacterCard({ character, language }: CharacterCardProps) {
  const title = character.title?.trim()
  const name = displayName(character)
  const portrait = character.assets.iconLarge || character.assets.iconMedium || character.iconPath

  return (
    <Link
      to="/characters/$characterId"
      params={{ characterId: String(character.id) }}
      search={{ lang: language }}
      className="group block rounded-[1rem] outline-none transition-transform duration-200 hover:-translate-y-0.5 focus-visible:ring-[3px] focus-visible:ring-ring/50"
    >
      <Card className="archive-card gap-0 overflow-hidden border-border/75 bg-card/82 p-2 py-2 shadow-none transition-colors group-hover:border-primary/45 group-hover:bg-card">
        <div className="archive-corner-frame relative overflow-hidden rounded-[0.82rem] border border-border/70 bg-secondary/50 p-1 shadow-[inset_0_1px_0_rgba(255,255,255,0.7)]">
          <CharacterPortrait
            key={portrait}
            character={character}
            src={portrait}
            className="block aspect-[4/5] w-full rounded-[0.58rem] bg-secondary object-cover object-top transition-transform duration-300 group-hover:scale-[1.018]"
          />
        </div>
        <CardContent className="flex flex-col gap-2 px-2 pb-2 pt-3">
          <div className="grid min-w-0 gap-1">
            <div className="flex items-baseline justify-between gap-2">
              <strong className="block truncate font-display text-[1.28rem] font-medium leading-tight tracking-[-0.025em] text-foreground" title={name}>
                {name}
              </strong>
              <span className="shrink-0 text-[0.68rem] tabular-nums text-primary/70">ID {character.id}</span>
            </div>
            {title ? <span className="line-clamp-2 text-[0.72rem] leading-5 text-muted-foreground">{title}</span> : null}
          </div>
          <div className="flex flex-wrap gap-1.5" aria-label="Character attributes">
            <AttributeIcon src={character.assets.elementIcon} label={formatElement(character.elementType)} />
            <AttributeIcon src={character.assets.jobIcon} label={formatJob(character.jobFlags)} />
          </div>
        </CardContent>
      </Card>
    </Link>
  )
}
