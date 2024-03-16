import { createLazyFileRoute } from "@tanstack/react-router";
import CharacterList from "../components/CharacterList";

function Index() {
  return (
    <>
      <CharacterList />
    </>
  );
}

export const Route = createLazyFileRoute("/")({
  component: Index,
});
