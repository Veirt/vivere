import * as React from "react";
import { ChevronDown, Search, SlidersHorizontal, X } from "lucide-react";

import { CharacterCard } from "#/components/character-card.tsx";
import { Button } from "#/components/ui/button.tsx";
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from "#/components/ui/collapsible.tsx";
import { Input } from "#/components/ui/input.tsx";
import { Select, SelectContent, SelectItem, SelectTrigger } from "#/components/ui/select.tsx";
import {
  filterCharacters,
  isSortKey,
  sortCharacters,
  sortOptions,
  type Character,
  type Language,
  type SortKey,
} from "#/lib/characters.ts";

const defaultSort: SortKey = "id";
const sortLabels: Record<SortKey, string> = {
  id: "ID",
  name: "Name",
  rarity: "Rarity",
  element: "Element",
};

type CharacterBrowserProps = {
  characters: ReadonlyArray<Character>;
  language: Language;
  query: string;
  sort: SortKey;
  onQueryChange: (query: string) => void;
  onSortChange: (sort: SortKey) => void;
};

export function CharacterBrowser({
  characters,
  language,
  query,
  sort,
  onQueryChange,
  onSortChange,
}: CharacterBrowserProps) {
  const [controlsOpen, setControlsOpen] = React.useState(false);
  const deferredQuery = React.useDeferredValue(query);
  const visibleCharacters = React.useMemo(
    () => sortCharacters(filterCharacters(characters, deferredQuery), sort),
    [characters, deferredQuery, sort],
  );
  const trimmedQuery = query.trim();
  const selectedSortLabel = sortLabels[sort];
  const hasCustomView = trimmedQuery.length > 0 || sort !== defaultSort;
  const status = trimmedQuery
    ? `${visibleCharacters.length} matches`
    : `${visibleCharacters.length} entries`;

  return (
    <section className="grid gap-5" id="characters" aria-labelledby="characters-title">
      <div className="archive-browser-hero relative isolate overflow-hidden rounded-[1.15rem] border border-border/75 bg-card/72 px-5 py-5 shadow-[0_18px_54px_rgba(101,72,38,0.06)] sm:px-7 lg:px-9">
        <div className="relative z-10 grid gap-5 lg:grid-cols-[minmax(0,1fr)_minmax(22rem,34rem)] lg:items-end">
          <div className="min-w-0">
            <div className="flex flex-wrap items-end gap-x-4 gap-y-1">
              <h1
                id="characters-title"
                className="font-display text-5xl font-normal leading-none tracking-[-0.055em] text-foreground sm:text-6xl"
              >
                Characters
              </h1>
              <p
                aria-hidden="true"
                className="archive-script -mb-1 text-4xl tracking-[0.04em] text-primary/28 sm:text-5xl"
              >
                Archive
              </p>
            </div>
            <p className="mt-3 max-w-[48ch] text-sm leading-7 text-muted-foreground">
              Search the illustrated records, then open a character folio for profile, lament,
              combat data, and local audio.
            </p>
          </div>

          <div className="grid gap-3">
            <label className="sr-only" htmlFor="character-search">
              Search characters
            </label>
            <div className="relative">
              <Search
                className="pointer-events-none absolute left-4 top-1/2 size-4 -translate-y-1/2 text-primary/70"
                strokeWidth={1.5}
                aria-hidden="true"
              />
              <Input
                id="character-search"
                value={query}
                onChange={(event) => onQueryChange(event.target.value)}
                placeholder="Search by name, title, skill, or story"
                className="h-12 rounded-[0.85rem] border-border/75 bg-background/45 pl-11 pr-11 font-display text-[1.05rem] shadow-[inset_0_1px_0_rgba(255,255,255,0.68)] placeholder:font-sans placeholder:text-sm"
              />
              {trimmedQuery ? (
                <Button
                  type="button"
                  variant="ghost"
                  size="icon-xs"
                  className="absolute right-3 top-1/2 -translate-y-1/2 rounded-full text-muted-foreground hover:bg-accent/45 hover:text-primary"
                  onClick={() => onQueryChange("")}
                  aria-label="Clear search"
                >
                  <X className="size-3.5" strokeWidth={1.6} aria-hidden="true" />
                </Button>
              ) : null}
            </div>

            <Collapsible open={controlsOpen} onOpenChange={setControlsOpen}>
              <div
                className="flex flex-wrap items-center justify-between gap-3 text-sm text-muted-foreground"
                role="status"
                aria-live="polite"
              >
                <span>{status}</span>
                <CollapsibleTrigger asChild>
                  <Button
                    type="button"
                    variant="outline"
                    size="sm"
                    className="archive-badge rounded-full border-border/75 bg-background/38 px-3 font-display text-foreground shadow-none hover:bg-accent/45"
                  >
                    <SlidersHorizontal
                      className="size-3.5 text-primary"
                      strokeWidth={1.5}
                      aria-hidden="true"
                    />
                    Refine
                    <ChevronDown
                      className={`size-3.5 transition-transform ${controlsOpen ? "rotate-180" : ""}`}
                      strokeWidth={1.5}
                      aria-hidden="true"
                    />
                  </Button>
                </CollapsibleTrigger>
              </div>

              <CollapsibleContent className="data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=closed]:slide-out-to-top-2 data-[state=open]:animate-in data-[state=open]:fade-in-0 data-[state=open]:slide-in-from-top-2">
                <div className="mt-3 grid gap-3 rounded-[0.9rem] border border-border/65 bg-background/35 p-3 shadow-[inset_0_1px_0_rgba(255,255,255,0.62)] sm:grid-cols-[minmax(0,1fr)_auto] sm:items-end">
                  <div className="grid gap-2">
                    <span
                      id="character-sort-label"
                      className="text-[0.68rem] font-semibold uppercase tracking-[0.18em] text-muted-foreground"
                    >
                      Arrangement
                    </span>
                    <Select
                      value={sort}
                      onValueChange={(value) => {
                        if (isSortKey(value)) {
                          onSortChange(value);
                        }
                      }}
                    >
                      <SelectTrigger
                        className="h-10 w-full rounded-[0.72rem] border-border/75 bg-card/55 font-display text-foreground shadow-none sm:w-[14rem]"
                        aria-labelledby="character-sort-label"
                      >
                        <span>{selectedSortLabel}</span>
                      </SelectTrigger>
                      <SelectContent>
                        {sortOptions.map((option) => (
                          <SelectItem key={option.value} value={option.value}>
                            {sortLabels[option.value]}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </div>

                  <Button
                    type="button"
                    variant="ghost"
                    size="sm"
                    className="justify-self-start rounded-full text-muted-foreground hover:bg-accent/45 hover:text-primary sm:justify-self-end"
                    disabled={!hasCustomView}
                    onClick={() => {
                      onQueryChange("");
                      onSortChange(defaultSort);
                    }}
                  >
                    Reset view
                  </Button>
                </div>
              </CollapsibleContent>
            </Collapsible>
          </div>
        </div>
      </div>

      <div aria-live="polite">
        {visibleCharacters.length > 0 ? (
          <div className="grid grid-cols-2 content-start gap-3 sm:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 2xl:grid-cols-6">
            {visibleCharacters.map((character) => (
              <CharacterCard key={character.id} character={character} language={language} />
            ))}
          </div>
        ) : (
          <div className="archive-card flex min-h-[25rem] flex-col justify-center rounded-[1rem] border border-border/75 bg-card/82 px-7 py-10 sm:p-14">
            <p className="mb-3 text-[0.68rem] font-semibold uppercase tracking-[0.22em] text-primary">
              No characters found
            </p>
            <h2 className="font-display text-[clamp(2rem,4vw,3rem)] font-normal tracking-[-0.04em] text-foreground">
              No folios match this search.
            </h2>
            <p className="mt-4 max-w-md text-sm leading-7 text-muted-foreground">
              Try a name, title, skill, or story phrase.
            </p>
            <Button
              type="button"
              variant="outline"
              size="sm"
              className="archive-badge mt-6 w-fit rounded-full bg-background/38 shadow-none"
              onClick={() => onQueryChange("")}
            >
              Clear search
            </Button>
          </div>
        )}
      </div>
    </section>
  );
}
