import { createFileRoute, useRouter, type ErrorComponentProps } from "@tanstack/react-router";
import { IconRefresh } from "@tabler/icons-react";
import { Effect } from "effect";

import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { VivereApi } from "../features/characters/api";
import { CharacterListView } from "../features/characters/components/CharacterListView";

function CharactersPending() {
  return (
    <main className="mx-auto max-w-4xl px-5 py-12 text-center">
      <p role="status" className="text-sm text-muted-foreground">
        Loading characters...
      </p>
    </main>
  );
}

function CharactersError({ error, reset }: ErrorComponentProps) {
  const router = useRouter();

  return (
    <main className="mx-auto max-w-lg px-5 py-12">
      <Card>
        <CardHeader className="text-center">
          <CardTitle className="text-xl">Unable to load characters</CardTitle>
          <CardDescription role="alert">
            {error instanceof Error ? error.message : "An unexpected error occurred."}
          </CardDescription>
        </CardHeader>
        <CardContent className="flex justify-center">
          <Button
            type="button"
            onClick={() => {
              reset();
              void router.invalidate();
            }}
          >
            <IconRefresh className="size-4" />
            Retry
          </Button>
        </CardContent>
      </Card>
    </main>
  );
}

function CharactersPage() {
  const characters = Route.useLoaderData();

  return <CharacterListView characters={characters} />;
}

export const Route = createFileRoute("/")({
  loader: ({ context, abortController }) =>
    context.runEffect(
      Effect.gen(function* () {
        const api = yield* VivereApi;

        return yield* api.listCharacters;
      }),
      abortController.signal,
    ),
  pendingComponent: CharactersPending,
  errorComponent: CharactersError,
  component: CharactersPage,
});
