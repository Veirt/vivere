import * as React from "react";
import { Crosshair, Droplet, Feather, Gauge, Sparkles, Star, type LucideIcon } from "lucide-react";

import { AttributeIcon } from "#/components/attribute-icon.tsx";
import { AudioList, AudioPreview } from "#/components/audio-preview.tsx";
import { CharacterPortrait } from "#/components/character-card.tsx";
import { CompactStatList, DetailGrid } from "#/components/detail-grid.tsx";
import { EmptyCard } from "#/components/empty-card.tsx";
import { Badge } from "#/components/ui/badge.tsx";
import { Card, CardContent, CardHeader, CardTitle } from "#/components/ui/card.tsx";
import { storyAudioSources, voiceAudioSources } from "#/lib/character-media.ts";
import {
  archiveQuote,
  cleanGameText,
  displayName,
  firstAvailableUrl,
  formatBirthday,
  formatBloodType,
  formatElement,
  formatHeight,
  formatJob,
  formatWeight,
  type Character,
  type CharacterSkill,
  type ParameterChange,
} from "#/lib/characters.ts";

const archiveCardClass = "archive-card gap-4 border-border/75 bg-card/82 py-5 shadow-none";
const archiveInsetClass =
  "rounded-[0.75rem] border border-border/65 bg-background/45 shadow-[inset_0_1px_0_rgba(255,255,255,0.66)]";

type DetailDatum = { label: string; value: React.ReactNode };

function SectionTitle({ children, level = 2 }: { children: React.ReactNode; level?: 1 | 2 | 3 }) {
  return (
    <CardTitle
      role="heading"
      aria-level={level}
      className="archive-section-title text-lg font-semibold tracking-[-0.01em] text-foreground"
    >
      <span>{children}</span>
    </CardTitle>
  );
}

function SmallCaps({ children }: { children: React.ReactNode }) {
  return (
    <span className="text-[0.68rem] font-semibold uppercase tracking-[0.18em] text-muted-foreground">
      {children}
    </span>
  );
}

function GlanceGlyph({ icon: Icon }: { icon: LucideIcon }) {
  return (
    <span className="archive-glance-glyph" aria-hidden="true">
      <Icon className="size-3.5" strokeWidth={1.45} />
    </span>
  );
}

function IconValue({ icon: Icon, children }: { icon: LucideIcon; children: React.ReactNode }) {
  return (
    <span className="inline-flex items-center gap-1.5 text-foreground">
      <Icon className="size-4 text-primary" strokeWidth={1.45} aria-hidden="true" />
      <span>{children}</span>
    </span>
  );
}

function SkillIcon({ skill }: { skill: CharacterSkill }) {
  const [failedSrc, setFailedSrc] = React.useState<string | null>(null);
  const usesMissingMasterIcon = /CSK_000000\d+\.png$/i.test(skill.icon || "");
  const hasImageError = failedSrc === skill.icon;

  const handleImageRef = React.useCallback(
    (image: HTMLImageElement | null) => {
      if (image?.complete && image.naturalWidth === 0) {
        setFailedSrc(skill.icon);
      }
    },
    [skill.icon],
  );

  if (!skill.icon || hasImageError || usesMissingMasterIcon) {
    return (
      <div className="flex size-14 items-center justify-center rounded-[0.65rem] border border-border/60 bg-secondary/58 text-primary">
        <Sparkles className="size-5" strokeWidth={1.45} aria-hidden="true" />
      </div>
    );
  }

  return (
    <img
      ref={handleImageRef}
      className="size-14 rounded-[0.65rem] border border-border/60 bg-secondary/70 object-contain p-1"
      src={skill.icon}
      alt=""
      loading="lazy"
      decoding="async"
      onLoad={(event) => {
        if (event.currentTarget.naturalWidth === 0) {
          setFailedSrc(skill.icon);
        }
      }}
      onError={() => setFailedSrc(skill.icon)}
    />
  );
}

export function CharacterArtworkCard({ character }: { character: Character }) {
  const detailImage =
    character.assets.iconLarge || character.assets.iconMedium || character.iconPath;

  return (
    <Card className={`${archiveCardClass} overflow-hidden p-2`}>
      <div className="archive-corner-frame relative overflow-hidden rounded-[0.9rem] border border-border/75 bg-secondary/55 p-1 shadow-[inset_0_1px_0_rgba(255,255,255,0.7)]">
        <CharacterPortrait
          key={detailImage}
          character={character}
          src={detailImage}
          className="block w-full rounded-[0.65rem] bg-secondary object-cover object-top"
        />
      </div>
    </Card>
  );
}

export function CharacterAtAGlanceCard({ character }: { character: Character }) {
  const metadata: Array<(DetailDatum & { icon: LucideIcon }) | null> = [
    {
      label: "Element",
      icon: Droplet,
      value: (
        <AttributeIcon
          src={character.assets.elementIcon}
          label={formatElement(character.elementType)}
          size="md"
          showLabel
        />
      ),
    },
    {
      label: "Class",
      icon: Crosshair,
      value: (
        <AttributeIcon
          src={character.assets.jobIcon}
          label={formatJob(character.jobFlags)}
          size="md"
          showLabel
        />
      ),
    },
    {
      label: "Rarity",
      icon: Star,
      value: <IconValue icon={Star}>Rank {character.rarityFlags}</IconValue>,
    },
    character.stats.speed
      ? {
          label: "Base speed",
          icon: Feather,
          value: <IconValue icon={Gauge}>{character.stats.speed.toLocaleString()}</IconValue>,
        }
      : null,
  ];

  return (
    <Card className={archiveCardClass}>
      <CardHeader className="px-5 pb-0">
        <SectionTitle>At a glance</SectionTitle>
      </CardHeader>
      <CardContent className="px-5">
        <dl className="grid overflow-hidden rounded-[0.85rem] border border-border/70 bg-background/35">
          {metadata
            .filter((detail): detail is DetailDatum & { icon: LucideIcon } => detail !== null)
            .map((detail) => (
              <div
                key={detail.label}
                className="grid grid-cols-[minmax(0,1fr)_auto] items-center gap-3 border-b border-border/55 px-4 py-2.5 last:border-b-0"
              >
                <dt className="inline-flex items-center gap-2 text-sm text-muted-foreground">
                  <GlanceGlyph icon={detail.icon} />
                  <span>{detail.label}</span>
                </dt>
                <dd className="text-sm font-medium tabular-nums text-foreground">{detail.value}</dd>
              </div>
            ))}
        </dl>
      </CardContent>
    </Card>
  );
}

export function CharacterAudioCard({
  character,
  autoplayAudio,
}: {
  character: Character;
  autoplayAudio: string;
}) {
  const sources = [
    character.assets.songShortJp
      ? { label: "Lament preview (JP)", src: character.assets.songShortJp }
      : null,
    character.assets.songShortUs
      ? { label: "Lament preview (US)", src: character.assets.songShortUs }
      : null,
    !character.assets.songShortJp && character.assets.songJp
      ? { label: "Lament (JP)", src: character.assets.songJp }
      : null,
    !character.assets.songShortUs && character.assets.songUs
      ? { label: "Lament (US)", src: character.assets.songUs }
      : null,
  ].filter((source): source is { label: string; src: string } => Boolean(source));

  return (
    <Card className={archiveCardClass}>
      <CardHeader className="px-5 pb-0">
        <SectionTitle>Audio</SectionTitle>
      </CardHeader>
      <CardContent className="grid gap-3 px-5">
        {sources.length > 0 ? (
          sources.map((source) => (
            <AudioPreview
              key={source.src}
              label={source.label}
              src={source.src}
              autoPlay={source.src === autoplayAudio}
            />
          ))
        ) : (
          <p className="text-sm leading-7 text-muted-foreground">
            No local audio assets are available for this character yet.
          </p>
        )}
      </CardContent>
    </Card>
  );
}

export function CharacterIdentityCard({ character }: { character: Character }) {
  const name = displayName(character);
  const quote = archiveQuote(character);

  return (
    <section
      className="archive-identity relative isolate overflow-hidden border-b border-border/75 px-5 py-6 sm:px-8 sm:py-8"
      aria-labelledby="character-identity-title"
    >
      <div className="archive-wash archive-wash-left" aria-hidden="true" />
      <div className="archive-wash archive-wash-right" aria-hidden="true" />

      <div className="relative z-10 grid gap-7 xl:grid-cols-[minmax(0,1fr)_minmax(18rem,33rem)] xl:items-center">
        <div className="min-w-0">
          <div className="mb-4 flex flex-wrap items-center gap-2">
            <Badge variant="secondary" className="archive-badge">
              ID {character.id}
            </Badge>
            <AttributeIcon
              src={character.assets.elementIcon}
              label={formatElement(character.elementType)}
              size="md"
              showLabel
            />
            <AttributeIcon
              src={character.assets.jobIcon}
              label={formatJob(character.jobFlags)}
              size="md"
              showLabel
            />
          </div>
          <div className="relative flex flex-wrap items-baseline gap-x-4 gap-y-2">
            <h1
              id="character-identity-title"
              className="relative z-10 font-display text-5xl font-normal leading-[0.95] tracking-[-0.055em] text-foreground sm:text-6xl lg:text-7xl"
            >
              {name}
            </h1>
            <p
              aria-hidden="true"
              className="archive-script pointer-events-none -mb-2 text-4xl tracking-[0.04em] text-primary/28 sm:text-5xl"
            >
              {name}
            </p>
          </div>
          {character.title ? (
            <p className="mt-2 font-display text-lg italic leading-7 text-primary/62">
              {cleanGameText(character.title)}
            </p>
          ) : null}
        </div>

        <figure className="justify-self-start text-primary/80 xl:justify-self-end">
          <blockquote className="archive-script archive-quote max-w-[28rem] text-3xl leading-[1.18] tracking-[0.02em] text-primary/78 sm:text-4xl">
            “{quote}”
          </blockquote>
          <figcaption className="mt-3 text-xs uppercase tracking-[0.28em] text-muted-foreground">
            archive note
          </figcaption>
        </figure>
      </div>
    </section>
  );
}

export function CharacterProfilePanel({ character }: { character: Character }) {
  const profileDetails = [
    character.profile.birthday
      ? { label: "Birthday", value: formatBirthday(character.profile.birthday) }
      : null,
    character.profile.bloodType
      ? { label: "Blood type", value: formatBloodType(character.profile.bloodType) }
      : null,
    character.profile.height
      ? { label: "Height", value: formatHeight(character.profile.height) }
      : null,
    character.profile.weight
      ? { label: "Weight", value: formatWeight(character.profile.weight) }
      : null,
    character.profile.voiceJp ? { label: "CV (JP)", value: character.profile.voiceJp } : null,
    character.profile.voiceUs ? { label: "CV (US)", value: character.profile.voiceUs } : null,
    character.profile.vocalJp ? { label: "Vocal (JP)", value: character.profile.vocalJp } : null,
    character.profile.vocalUs ? { label: "Vocal (US)", value: character.profile.vocalUs } : null,
  ].filter((detail): detail is { label: string; value: string } => Boolean(detail));

  return (
    <Card id="overview" className={`${archiveCardClass} archive-panel-botanical scroll-mt-28`}>
      <CardHeader className="px-5 pb-0 sm:px-6">
        <SectionTitle>Profile</SectionTitle>
      </CardHeader>
      <CardContent className="grid gap-6 px-5 sm:px-6 lg:grid-cols-[minmax(0,1fr)_minmax(18rem,38rem)] lg:items-center">
        <div className="space-y-4">
          {character.profile.description ? (
            <p className="max-w-[76ch] text-[0.98rem] leading-8 text-foreground/88">
              {cleanGameText(character.profile.description)}
            </p>
          ) : null}
        </div>

        {profileDetails.length > 0 ? <DetailGrid details={profileDetails} /> : null}
      </CardContent>
    </Card>
  );
}

export function CharacterStatsPanel({ character }: { character: Character }) {
  const battleStats = Object.entries(character.stats.initialBattleParameter).filter(
    ([, value]) => value !== 0,
  );
  const coefficients = Object.entries(character.stats.baseParameterCoefficient).filter(
    ([, value]) => value !== 0,
  );

  return (
    <Card id="stats" className={`${archiveCardClass} archive-panel-crest scroll-mt-28`}>
      <CardHeader className="px-5 pb-0 sm:px-6">
        <SectionTitle>Combat profile</SectionTitle>
      </CardHeader>
      <CardContent className="grid gap-5 px-5 sm:px-6">
        <div className="flex flex-wrap gap-2">
          {character.stats.normalSkillId ? (
            <Badge variant="secondary" className="archive-badge">
              Normal {character.stats.normalSkillId}
            </Badge>
          ) : null}
          {character.stats.activeSkillIds.map((id) => (
            <Badge key={`active-${id}`} variant="outline" className="archive-badge">
              Active {id}
            </Badge>
          ))}
          {character.stats.passiveSkillIds.map((id) => (
            <Badge key={`passive-${id}`} variant="outline" className="archive-badge">
              Passive {id}
            </Badge>
          ))}
        </div>
        {coefficients.length > 0 ? (
          <CompactStatList title="Base coefficients" entries={coefficients} />
        ) : null}
        {battleStats.length > 0 ? (
          <CompactStatList title="Initial battle modifiers" entries={battleStats} />
        ) : null}
      </CardContent>
    </Card>
  );
}

export function CharacterSkillsPanel({ character }: { character: Character }) {
  const skills = [
    character.skills.normal.id ? character.skills.normal : null,
    ...character.skills.active,
    ...character.skills.passive,
  ].filter((skill): skill is Character["skills"]["normal"] => Boolean(skill?.id));

  if (skills.length === 0) {
    return (
      <EmptyCard
        title="Skills"
        description="No detailed skill data is available for this character."
      />
    );
  }

  return (
    <Card id="skills" className={`${archiveCardClass} scroll-mt-28`}>
      <CardHeader className="px-5 pb-0 sm:px-6">
        <div className="flex flex-wrap items-baseline gap-x-5 gap-y-2">
          <SectionTitle>Skills</SectionTitle>
          <p className="text-sm text-muted-foreground">
            Active and passive skills translated from master data.
          </p>
        </div>
      </CardHeader>
      <CardContent className="grid gap-3 px-5 sm:px-6">
        {skills.map((skill) => (
          <SkillCard key={`${skill.type}-${skill.id}`} skill={skill} />
        ))}
      </CardContent>
    </Card>
  );
}

export function CharacterUniqueWeaponPanel({ character }: { character: Character }) {
  const weapon = character.uniqueWeapon;
  if (!weapon.equipmentId) {
    return (
      <EmptyCard
        title="Unique Weapon"
        description="This character does not have unique weapon data yet."
      />
    );
  }

  const descriptions = [
    weapon.descriptions.ssr ? { label: "SSR", value: weapon.descriptions.ssr } : null,
    weapon.descriptions.ur ? { label: "UR", value: weapon.descriptions.ur } : null,
    weapon.descriptions.lr ? { label: "LR", value: weapon.descriptions.lr } : null,
  ].filter((description): description is { label: string; value: string } => Boolean(description));

  return (
    <Card id="weapon" className={`${archiveCardClass} scroll-mt-28`}>
      <CardHeader className="px-5 pb-0 sm:px-6">
        <SectionTitle>Unique Weapon</SectionTitle>
      </CardHeader>
      <CardContent className="grid gap-5 px-5 sm:px-6">
        <div className={`flex gap-4 ${archiveInsetClass} p-4`}>
          {weapon.icon ? (
            <img
              className="size-16 rounded-[0.65rem] border border-border/70 bg-secondary/70 object-contain p-1"
              src={weapon.icon}
              alt=""
              loading="lazy"
              decoding="async"
            />
          ) : null}
          <div className="min-w-0 space-y-2">
            <h3 className="font-display text-xl font-medium leading-tight text-foreground">
              {weapon.name || weapon.nameKey}
            </h3>
            <div className="flex flex-wrap gap-2">
              <Badge variant="secondary" className="archive-badge">
                Lv {weapon.level}
              </Badge>
              <Badge variant="outline" className="archive-badge">
                Rarity {weapon.rarityFlags}
              </Badge>
              {weapon.jobFlags ? (
                <Badge variant="outline" className="archive-badge">
                  {formatJob(weapon.jobFlags)}
                </Badge>
              ) : null}
            </div>
          </div>
        </div>

        {descriptions.length > 0 ? (
          <div className="grid gap-3">
            {descriptions.map((description) => (
              <section
                key={description.label}
                className={`${archiveInsetClass} p-4`}
                aria-label={`${description.label} weapon bonus`}
              >
                <h3 className="mb-2">
                  <SmallCaps>{description.label} Weapon</SmallCaps>
                </h3>
                <p className="whitespace-pre-wrap text-sm leading-7 text-muted-foreground">
                  {cleanGameText(description.value)}
                </p>
              </section>
            ))}
          </div>
        ) : null}

        {weapon.bonusParameters.length > 0 ? (
          <WeaponBonusList bonuses={weapon.bonusParameters} />
        ) : null}
      </CardContent>
    </Card>
  );
}

function SkillCard({ skill }: { skill: CharacterSkill }) {
  return (
    <article className={`${archiveInsetClass} p-4`}>
      <div className="flex gap-4">
        <SkillIcon skill={skill} />
        <div className="min-w-0 flex-1">
          <div className="flex flex-wrap items-center gap-2">
            <h3 className="font-medium text-foreground">
              {skill.name || skill.nameKey || `Skill ${skill.id}`}
            </h3>
            <Badge
              variant={skill.type === "passive" ? "secondary" : "outline"}
              className="archive-badge"
            >
              {skill.type}
            </Badge>
            {skill.maxCooltime ? (
              <Badge variant="outline" className="archive-badge">
                CD {skill.maxCooltime}
              </Badge>
            ) : null}
          </div>
          {skill.conditions && skill.conditions !== "True" ? (
            <p className="mt-1 text-xs text-muted-foreground">Condition: {skill.conditions}</p>
          ) : null}
        </div>
      </div>
      {skill.levels.length > 0 ? (
        <div className="mt-4 grid gap-3">
          {skill.levels.map((level) => (
            <section
              key={`${skill.id}-${level.orderNumber}-${level.characterLevel}-${level.equipmentRarityFlags}`}
              className="rounded-[0.7rem] border border-border/60 bg-card/55 p-3"
            >
              <div className="mb-2 flex flex-wrap gap-2">
                {level.orderNumber ? (
                  <Badge variant="outline" className="archive-badge">
                    Lv {level.orderNumber}
                  </Badge>
                ) : null}
                {level.characterLevel ? (
                  <Badge variant="secondary" className="archive-badge">
                    Character {level.characterLevel}
                  </Badge>
                ) : null}
                {level.equipmentRarityFlags ? (
                  <Badge variant="secondary" className="archive-badge">
                    Weapon {level.equipmentRarityFlags}
                  </Badge>
                ) : null}
              </div>
              {level.description ? (
                <p className="whitespace-pre-wrap text-sm leading-7 text-muted-foreground">
                  {cleanGameText(level.description)}
                </p>
              ) : null}
            </section>
          ))}
        </div>
      ) : null}
    </article>
  );
}

function WeaponBonusList({ bonuses }: { bonuses: ReadonlyArray<ParameterChange> }) {
  const seenCounts = new Map<string, number>();
  const keyedBonuses = bonuses.map((bonus) => {
    const baseKey = `${bonus.category}-${bonus.type}-${bonus.changeType}-${bonus.value}`;
    const count = (seenCounts.get(baseKey) ?? 0) + 1;
    seenCounts.set(baseKey, count);
    return { bonus, key: count === 1 ? baseKey : `${baseKey}#${count}` };
  });

  return (
    <section aria-label="Unique weapon bonus parameters">
      <h3 className="mb-3">
        <SmallCaps>Bonus parameters</SmallCaps>
      </h3>
      <dl className="grid gap-2 sm:grid-cols-2">
        {keyedBonuses.map(({ bonus, key }) => (
          <div
            key={key}
            className={`flex items-center justify-between gap-3 ${archiveInsetClass} px-3 py-2`}
          >
            <dt className="text-xs text-muted-foreground">
              {bonus.category} #{bonus.type}
            </dt>
            <dd className="text-sm font-medium tabular-nums text-foreground">
              {bonus.value.toLocaleString()}
            </dd>
          </div>
        ))}
      </dl>
    </section>
  );
}

export function CharacterLamentPanel({ character }: { character: Character }) {
  const jpMovie = firstAvailableUrl(character.profile.movieJpUrls, ["jaJP", "enUS"]);
  const usMovie = firstAvailableUrl(character.profile.movieUsUrls, ["enUS", "jaJP"]);
  const hasLament =
    character.profile.lamentJp ||
    character.profile.lamentUs ||
    character.profile.lyricsJp ||
    character.profile.lyricsUs ||
    jpMovie ||
    usMovie;

  if (!hasLament) {
    return (
      <EmptyCard title="Lament" description="No lament data is available for this character." />
    );
  }

  return (
    <Card id="lament" className={`${archiveCardClass} archive-panel-botanical scroll-mt-28`}>
      <CardHeader className="px-5 pb-0 sm:px-6">
        <SectionTitle>Lament</SectionTitle>
      </CardHeader>
      <CardContent className="grid gap-4 px-5 sm:px-6">
        <div className="flex flex-wrap gap-2">
          {character.profile.lamentJp ? (
            <Badge variant="secondary" className="archive-badge">
              JP: {character.profile.lamentJp}
            </Badge>
          ) : null}
          {character.profile.lamentUs ? (
            <Badge variant="secondary" className="archive-badge">
              US: {character.profile.lamentUs}
            </Badge>
          ) : null}
          {jpMovie ? (
            <Badge asChild variant="outline" className="archive-badge">
              <a href={jpMovie}>YouTube JP</a>
            </Badge>
          ) : null}
          {usMovie ? (
            <Badge asChild variant="outline" className="archive-badge">
              <a href={usMovie}>YouTube US</a>
            </Badge>
          ) : null}
        </div>
        <div className="grid gap-3 lg:grid-cols-2">
          {character.profile.lyricsJp ? (
            <LyricsBlock title="Japanese lyrics" text={character.profile.lyricsJp} />
          ) : null}
          {character.profile.lyricsUs ? (
            <LyricsBlock title="English lyrics" text={character.profile.lyricsUs} />
          ) : null}
        </div>
      </CardContent>
    </Card>
  );
}

export function CharacterStoriesPanel({
  stories,
  characterName,
}: {
  stories: Character["stories"];
  characterName: string;
}) {
  if (stories.length === 0) {
    return (
      <EmptyCard
        title="Memories"
        description={`No story episodes are available for ${characterName}.`}
      />
    );
  }

  return (
    <Card id="memories" className={`${archiveCardClass} scroll-mt-28`}>
      <CardHeader className="px-5 pb-0 sm:px-6">
        <div className="flex flex-wrap items-baseline gap-x-5 gap-y-2">
          <SectionTitle>Memories</SectionTitle>
          <p className="text-sm text-muted-foreground">
            Story episodes and local voice files for {characterName}.
          </p>
        </div>
      </CardHeader>
      <CardContent className="grid gap-3 px-5 sm:px-6">
        {stories.map((story) => (
          <article key={story.id} className={`${archiveInsetClass} p-4`}>
            <div className="mb-2 flex flex-wrap items-center gap-2">
              <Badge variant="outline" className="archive-badge">
                Episode {story.episodeId}
              </Badge>
              {story.level ? (
                <Badge variant="secondary" className="archive-badge">
                  Level {story.level}
                </Badge>
              ) : null}
              {story.rarityFlags ? (
                <Badge variant="secondary" className="archive-badge">
                  Rarity {story.rarityFlags}
                </Badge>
              ) : null}
            </div>
            <h3 className="font-display text-lg font-medium text-foreground">
              {story.title || story.titleKey}
            </h3>
            {story.text ? (
              <p className="mt-2 whitespace-pre-wrap text-sm leading-7 text-muted-foreground">
                {cleanGameText(story.text)}
              </p>
            ) : null}
            <AudioList title="Story audio" sources={storyAudioSources(story)} />
          </article>
        ))}
      </CardContent>
    </Card>
  );
}

export function CharacterVoicesPanel({
  voices,
  characterName,
}: {
  voices: Character["voices"];
  characterName: string;
}) {
  const visibleVoices = voices.filter(
    (voice) => voice.subtitle || voice.buttonText || voice.audioJp || voice.audioUs,
  );
  if (visibleVoices.length === 0) {
    return (
      <EmptyCard
        title="Voicelines"
        description={`No voicelines are available for ${characterName}.`}
      />
    );
  }

  return (
    <Card id="voices" className={`${archiveCardClass} scroll-mt-28`}>
      <CardHeader className="px-5 pb-0 sm:px-6">
        <div className="flex flex-wrap items-baseline gap-x-5 gap-y-2">
          <SectionTitle>Voicelines</SectionTitle>
          <p className="text-sm text-muted-foreground">
            Localized subtitles, unlock labels, and local voice files for {characterName}.
          </p>
        </div>
      </CardHeader>
      <CardContent className="grid gap-3 px-5 sm:px-6 md:grid-cols-2">
        {visibleVoices.map((voice) => (
          <article key={voice.id} className={`${archiveInsetClass} p-4`}>
            <div className="mb-2 flex flex-wrap items-center gap-2">
              <Badge variant="outline" className="archive-badge">
                #{voice.sortOrder || voice.id}
              </Badge>
              {voice.category ? (
                <Badge variant="secondary" className="archive-badge">
                  Category {voice.category}
                </Badge>
              ) : null}
            </div>
            <h3 className="text-sm font-medium text-foreground">
              {voice.buttonText || voice.buttonTextKey}
            </h3>
            {voice.subtitle ? (
              <p className="mt-2 text-sm leading-7 text-muted-foreground">
                {cleanGameText(voice.subtitle)}
              </p>
            ) : null}
            <AudioList title="Voiceline audio" sources={voiceAudioSources(voice)} />
          </article>
        ))}
      </CardContent>
    </Card>
  );
}

function LyricsBlock({ title, text }: { title: string; text: string }) {
  return (
    <section className={`${archiveInsetClass} p-4`} aria-label={title}>
      <h3 className="mb-3">
        <SmallCaps>{title}</SmallCaps>
      </h3>
      <p className="line-clamp-12 whitespace-pre-wrap text-sm leading-7 text-muted-foreground">
        {cleanGameText(text)}
      </p>
    </section>
  );
}
