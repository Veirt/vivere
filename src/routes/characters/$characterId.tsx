import { createFileRoute } from "@tanstack/react-router";
import CharacterDetail from "../../components/CharacterDetail";

function CharacterPage() {
  const { characterId } = Route.useParams();

  return <CharacterDetail id={characterId} />;
}

export const Route = createFileRoute("/characters/$characterId")({
  component: CharacterPage,
});
