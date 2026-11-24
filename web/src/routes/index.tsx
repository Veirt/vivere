import { useSuspenseQuery } from "@tanstack/react-query";
import { createFileRoute, type ErrorComponentProps } from "@tanstack/react-router";
import type * as React from "react";

import { CharacterBrowser } from "#/components/character-browser.tsx";
import { SiteHeader } from "#/components/site-header.tsx";
import { Card, CardContent, CardHeader, CardTitle } from "#/components/ui/card.tsx";
import { Skeleton } from "#/components/ui/skeleton.tsx";
import {
  type CharacterIndexSearch,
  type Language,
  charactersQueryOptions,
  validateCharacterIndexSearch,
} from "#/lib/characters.ts";

const PENDING_CARD_KEYS = Array.from(
  { length: 18 },
  (_, position) => `pending-card-${position + 1}`,
);

export const Route = createFileRoute("/")({
  validateSearch: validateCharacterIndexSearch,
  loaderDeps: ({ search }) => ({ lang: search.lang }),
  loader: ({ context, deps }) =>
    context.queryClient.ensureQueryData(charactersQueryOptions(deps.lang)),
  pendingComponent: IndexPending,
  errorComponent: IndexError,
  component: RouteComponent,
});

function RouteComponent() {
  const search = Route.useSearch();
  const { data: characters } = useSuspenseQuery(charactersQueryOptions(search.lang));
  const navigate = Route.useNavigate();

  const updateSearch = (nextSearch: Partial<CharacterIndexSearch>) => {
    void navigate({
      search: (previous) => ({ ...previous, ...nextSearch }),
      replace: true,
    });
  };

  return (
    <IndexScaffold language={search.lang} onLanguageChange={(lang) => updateSearch({ lang })}>
      <CharacterBrowser
        characters={characters}
        language={search.lang}
        query={search.q}
        sort={search.sort}
        onQueryChange={(q) => updateSearch({ q })}
        onSortChange={(sort) => updateSearch({ sort })}
      />
    </IndexScaffold>
  );
}

function IndexPending() {
  const search = Route.useSearch();
  const navigate = Route.useNavigate();

  return (
    <IndexScaffold
      language={search.lang}
      onLanguageChange={(lang) => {
        void navigate({ search: (previous) => ({ ...previous, lang }), replace: true });
      }}
    >
      <section className="grid gap-5" aria-label="Loading characters" role="status">
        <Card className="archive-card gap-4 border-border/75 bg-card/82 py-5 shadow-none">
          <CardHeader>
            <Skeleton className="h-5 w-24" aria-hidden="true" />
            <Skeleton className="h-12 w-64" aria-hidden="true" />
            <Skeleton className="h-12 w-full max-w-xl" aria-hidden="true" />
          </CardHeader>
        </Card>
        <div className="grid grid-cols-2 content-start gap-3 sm:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 2xl:grid-cols-6">
          {PENDING_CARD_KEYS.map((key) => (
            <Skeleton key={key} className="aspect-[0.72] rounded-[1rem]" aria-hidden="true" />
          ))}
        </div>
      </section>
    </IndexScaffold>
  );
}

function IndexError({ error }: ErrorComponentProps) {
  const search = Route.useSearch();
  const navigate = Route.useNavigate();
  const message = error instanceof Error ? error.message : "Failed to load characters";

  return (
    <IndexScaffold
      language={search.lang}
      onLanguageChange={(lang) => {
        void navigate({ search: (previous) => ({ ...previous, lang }), replace: true });
      }}
    >
      <Card
        className="archive-card gap-4 border-border/75 bg-card/82 py-5 shadow-none"
        role="alert"
      >
        <CardHeader>
          <CardTitle>Could not load characters</CardTitle>
        </CardHeader>
        <CardContent>
          <p className="text-sm text-muted-foreground">{message}</p>
        </CardContent>
      </Card>
    </IndexScaffold>
  );
}

type IndexScaffoldProps = {
  children: React.ReactNode;
  language: Language;
  onLanguageChange: (language: Language) => void;
};

function IndexScaffold({ children, language, onLanguageChange }: IndexScaffoldProps) {
  return (
    <main className="archive-page min-h-[100dvh] bg-background px-3 pb-6 pt-3 text-foreground sm:px-6 sm:pb-10 sm:pt-5">
      <div className="archive-folio-shell mx-auto w-full max-w-[2440px] overflow-hidden rounded-[1.35rem] border border-border/75 shadow-[0_22px_72px_rgba(90,61,31,0.09)]">
        <SiteHeader language={language} onLanguageChange={onLanguageChange} />
        <div className="px-4 py-5 sm:px-7 sm:py-7">{children}</div>
      </div>
    </main>
  );
}
