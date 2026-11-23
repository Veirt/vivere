import { Link } from "@tanstack/react-router";
import { Array } from "effect";

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import type { CharacterSummary } from "../domain";

export interface CharacterListViewProps {
  readonly characters: ReadonlyArray<CharacterSummary>;
}

export function CharacterListView({ characters }: CharacterListViewProps) {
  if (characters.length === 0) {
    return (
      <main className="mx-auto max-w-lg px-5 py-12">
        <Card>
          <CardHeader className="text-center">
            <CardTitle className="text-xl">No characters found</CardTitle>
            <CardDescription>
              Start the asset server with Live2D models available to browse characters.
            </CardDescription>
          </CardHeader>
        </Card>
      </main>
    );
  }

  return (
    <main className="mx-auto max-w-6xl px-5 py-6">
      <h1 className="sr-only">Characters</h1>
      <div className="flex flex-wrap items-center justify-center gap-5 my-5">
        {Array.map(characters, (character) => (
          <Link
            key={character.id}
            to="/characters/$characterId"
            params={{ characterId: character.id }}
            aria-label={`Inspect character ${character.id}`}
            className="group rounded-xl transition-transform hover:scale-105 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
          >
            <Card size="sm" className="transition-shadow group-hover:shadow-sm">
              <CardContent className="p-2">
                <img
                  src={character.spriteSPath}
                  alt={`${character.id} portrait`}
                  className="block rounded-lg"
                />
              </CardContent>
            </Card>
          </Link>
        ))}
      </div>
    </main>
  );
}
