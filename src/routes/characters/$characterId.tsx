import { createFileRoute, Link, useRouter, type ErrorComponentProps } from "@tanstack/react-router";
import { IconArrowLeft, IconRefresh } from "@tabler/icons-react";
import { Effect, Schema } from "effect";

import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { VivereApi } from "../../features/characters/api";
import { CharacterDetailView } from "../../features/characters/components/CharacterDetailView";
import { CharacterId } from "../../features/characters/domain";
import { CharacterNotFoundError } from "../../features/characters/errors";

function CharacterDetailPending() {
  return (
    <main className="mx-auto max-w-4xl px-5 py-12 text-center">
      <p role="status" className="text-sm text-muted-foreground">
        Loading character profile...
      </p>
    </main>
  );
}

function CharacterDetailError({ error, reset }: ErrorComponentProps) {
  const router = useRouter();
  const isNotFound = Schema.is(CharacterNotFoundError)(error);

  return (
    <main className="mx-auto max-w-lg px-5 py-12">
      <Card>
        <CardHeader className="text-center">
          <CardTitle className="text-xl">
            {isNotFound ? "Character not found" : "Unable to load character"}
          </CardTitle>
          <CardDescription role="alert">
            {isNotFound
              ? `No character matched "${error.characterId}".`
              : error instanceof Error
                ? error.message
                : "An unexpected error occurred."}
          </CardDescription>
        </CardHeader>
        <CardContent className="flex justify-center gap-3">
          <Button variant="outline" asChild>
            <Link to="/">
              <IconArrowLeft className="size-4" />
              Back to characters
            </Link>
          </Button>
          {!isNotFound ? (
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
          ) : null}
        </CardContent>
      </Card>
    </main>
  );
}

function CharacterPage() {
  const detail = Route.useLoaderData();

  return <CharacterDetailView key={detail.id} detail={detail} />;
}

export const Route = createFileRoute("/characters/$characterId")({
  loader: ({ context, params, abortController }) =>
    context.runEffect(
      Effect.gen(function* () {
        const id = yield* Schema.decodeEffect(CharacterId)(params.characterId).pipe(
          Effect.tapError((error) => Effect.logWarning("Invalid characterId route param", error)),
          Effect.mapError(() => CharacterNotFoundError.make({ characterId: params.characterId })),
        );

        const api = yield* VivereApi;

        return yield* api.getCharacterDetail({ id });
      }),
      abortController.signal,
    ),
  pendingComponent: CharacterDetailPending,
  errorComponent: CharacterDetailError,
  component: CharacterPage,
});
