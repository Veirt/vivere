import { useSuspenseQuery } from "@tanstack/react-query";
import { Link, createFileRoute, type ErrorComponentProps } from "@tanstack/react-router";
import type * as React from "react";

import {
  CharacterArtworkCard,
  CharacterAtAGlanceCard,
  CharacterAudioCard,
  CharacterIdentityCard,
  CharacterLamentPanel,
  CharacterProfilePanel,
  CharacterSkillsPanel,
  CharacterStatsPanel,
  CharacterUniqueWeaponPanel,
  CharacterStoriesPanel,
  CharacterVoicesPanel,
} from "#/components/character-detail-panels.tsx";
import { SiteHeader } from "#/components/site-header.tsx";
import { Card, CardContent, CardHeader, CardTitle } from "#/components/ui/card.tsx";
import { Skeleton } from "#/components/ui/skeleton.tsx";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "#/components/ui/tabs.tsx";
import { TooltipProvider } from "#/components/ui/tooltip.tsx";
import { preferredAutoplayAudio } from "#/lib/character-media.ts";
import {
  characterQueryOptions,
  displayName,
  validateCharacterDetailSearch,
  type Character,
  type Language,
} from "#/lib/characters.ts";

export const Route = createFileRoute("/characters/$characterId")({
  validateSearch: validateCharacterDetailSearch,
  loaderDeps: ({ search }) => ({ lang: search.lang }),
  loader: ({ context, params, deps }) => {
    const characterId = Number(params.characterId);

    if (!Number.isInteger(characterId)) {
      return Promise.reject(new Error("Invalid character id"));
    }

    return context.queryClient.ensureQueryData(characterQueryOptions(characterId, deps.lang));
  },
  pendingComponent: CharacterDetailPending,
  errorComponent: CharacterDetailError,
  component: CharacterDetailRoute,
});

function CharacterDetailRoute() {
  const { lang } = Route.useSearch();
  const { characterId } = Route.useParams();
  const { data: character } = useSuspenseQuery(characterQueryOptions(Number(characterId), lang));
  const navigate = Route.useNavigate();

  return (
    <CharacterDetailScaffold
      language={lang}
      onLanguageChange={(nextLanguage) => {
        void navigate({ search: (previous) => ({ ...previous, lang: nextLanguage }) });
      }}
    >
      {character === null ? (
        <CharacterDetailMessage
          title="Character not found"
          description="Try another character from the list."
          role="alert"
        />
      ) : (
        <CharacterDetail character={character} />
      )}
    </CharacterDetailScaffold>
  );
}

function CharacterDetailPending() {
  const { lang } = Route.useSearch();
  const navigate = Route.useNavigate();

  return (
    <CharacterDetailScaffold
      language={lang}
      onLanguageChange={(nextLanguage) => {
        void navigate({ search: (previous) => ({ ...previous, lang: nextLanguage }) });
      }}
    >
      <CharacterDetailSkeleton />
    </CharacterDetailScaffold>
  );
}

function CharacterDetailError({ error }: ErrorComponentProps) {
  const { lang } = Route.useSearch();
  const navigate = Route.useNavigate();
  const message = error instanceof Error ? error.message : "Failed to load character";

  return (
    <CharacterDetailScaffold
      language={lang}
      onLanguageChange={(nextLanguage) => {
        void navigate({ search: (previous) => ({ ...previous, lang: nextLanguage }) });
      }}
    >
      <CharacterDetailMessage title="Could not load character" description={message} role="alert" />
    </CharacterDetailScaffold>
  );
}

type CharacterDetailScaffoldProps = {
  children: React.ReactNode;
  language: Language;
  onLanguageChange: (language: Language) => void;
};

function CharacterDetailScaffold({
  children,
  language,
  onLanguageChange,
}: CharacterDetailScaffoldProps) {
  return (
    <main className="archive-page min-h-[100dvh] bg-background px-3 pb-6 pt-3 text-foreground sm:px-6 sm:pb-10 sm:pt-5">
      <div className="archive-folio-shell mx-auto w-full max-w-[2440px] overflow-hidden rounded-[1.35rem] border border-border/75 shadow-[0_22px_72px_rgba(90,61,31,0.09)]">
        <SiteHeader language={language} onLanguageChange={onLanguageChange} />

        <div className="px-4 py-5 sm:px-7 sm:py-7">
          <Link
            to="/"
            search={{ lang: language, q: "", sort: "id" }}
            className="inline-flex items-center gap-2 font-display text-sm text-primary underline-offset-4 transition hover:text-foreground hover:underline"
          >
            <span aria-hidden="true">‹</span> Back to characters
          </Link>

          {children}
        </div>
      </div>
    </main>
  );
}

export function CharacterDetail({ character }: { character: Character }) {
  const name = displayName(character);
  const autoplayAudio = preferredAutoplayAudio(character);

  return (
    <TooltipProvider>
      <section className="mt-4 grid gap-5 xl:grid-cols-[minmax(20rem,30rem)_minmax(0,1fr)]">
        <aside className="archive-sidebar grid content-start gap-5">
          <CharacterArtworkCard character={character} />
          <CharacterAtAGlanceCard character={character} />
          <CharacterAudioCard character={character} autoplayAudio={autoplayAudio} />
        </aside>

        <div className="min-w-0 overflow-hidden rounded-[1.15rem] border border-border/75 bg-card/72 shadow-[0_18px_54px_rgba(101,72,38,0.06)]">
          <CharacterIdentityCard character={character} />

          <Tabs defaultValue="overview" className="min-w-0 gap-0">
            <TabsList
              aria-label="Character detail sections"
              className="archive-tabs rounded-none border-0 border-b border-border/75 bg-transparent px-4 py-0 sm:px-7"
            >
              <TabsTrigger value="overview">Overview</TabsTrigger>
              <TabsTrigger value="skills">Skills</TabsTrigger>
              <TabsTrigger value="weapon">Unique Weapon</TabsTrigger>
              <TabsTrigger value="lament">Lament</TabsTrigger>
              <TabsTrigger value="memories">Memories</TabsTrigger>
              <TabsTrigger value="voices">Voicelines</TabsTrigger>
              <TabsTrigger value="stats">Stats</TabsTrigger>
            </TabsList>

            <TabsContent value="overview" className="p-3 sm:p-4">
              <CharacterOverviewPanels character={character} characterName={name} />
            </TabsContent>
            <TabsContent value="skills" className="p-3 sm:p-4">
              <CharacterSkillsPanel character={character} />
            </TabsContent>
            <TabsContent value="weapon" className="p-3 sm:p-4">
              <CharacterUniqueWeaponPanel character={character} />
            </TabsContent>
            <TabsContent value="lament" className="p-3 sm:p-4">
              <CharacterLamentPanel character={character} />
            </TabsContent>
            <TabsContent value="memories" className="p-3 sm:p-4">
              <CharacterStoriesPanel stories={character.stories} characterName={name} />
            </TabsContent>
            <TabsContent value="voices" className="p-3 sm:p-4">
              <CharacterVoicesPanel voices={character.voices} characterName={name} />
            </TabsContent>
            <TabsContent value="stats" className="p-3 sm:p-4">
              <CharacterStatsPanel character={character} />
            </TabsContent>
          </Tabs>
        </div>
      </section>
    </TooltipProvider>
  );
}

function CharacterOverviewPanels({ character }: { character: Character; characterName: string }) {
  return (
    <div className="grid gap-5 2xl:grid-cols-[minmax(0,1fr)_minmax(24rem,0.92fr)]">
      <div className="2xl:col-span-2">
        <CharacterProfilePanel character={character} />
      </div>
      <div className="grid content-start gap-5">
        <CharacterSkillsPanel character={character} />
        <CharacterUniqueWeaponPanel character={character} />
      </div>
      <div className="grid content-start gap-5">
        <CharacterLamentPanel character={character} />
        <CharacterStatsPanel character={character} />
      </div>
    </div>
  );
}

function CharacterDetailSkeleton() {
  return (
    <section
      className="mt-4 grid gap-5 xl:grid-cols-[minmax(18rem,28rem)_minmax(0,1fr)]"
      aria-label="Loading character detail"
      role="status"
    >
      <Skeleton className="rounded-[1.25rem]" aria-hidden="true" />
      <Card className="archive-card border-border/80 bg-card/82">
        <CardHeader>
          <Skeleton className="h-4 w-24" aria-hidden="true" />
          <Skeleton className="h-10 w-2/3" aria-hidden="true" />
          <Skeleton className="h-4 w-1/2" aria-hidden="true" />
        </CardHeader>
        <CardContent className="grid gap-3">
          <Skeleton className="h-4 w-full" aria-hidden="true" />
          <Skeleton className="h-4 w-5/6" aria-hidden="true" />
          <Skeleton className="h-4 w-2/3" aria-hidden="true" />
        </CardContent>
      </Card>
    </section>
  );
}

function CharacterDetailMessage({
  title,
  description,
  role,
}: {
  title: string;
  description: string;
  role?: "alert";
}) {
  return (
    <Card
      className="archive-card mt-6 border-border/80 bg-card/82"
      {...(role !== undefined ? { role } : {})}
    >
      <CardHeader>
        <CardTitle role="heading" aria-level={2}>
          {title}
        </CardTitle>
      </CardHeader>
      <CardContent>
        <p className="text-sm text-muted-foreground">{description}</p>
      </CardContent>
    </Card>
  );
}
