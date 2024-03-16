import { Link } from "@tanstack/react-router";
import { useEffect, useState } from "react";

type Character = {
  id: string;
  spriteSPath: string;
};

export default function CharacterList() {
  const [characters, setCharacters] = useState<Character[]>([]);
  useEffect(() => {
    async function fetchCharacters() {
      const res = await fetch("/api/characters");
      const data = await res.json();

      setCharacters(data.characters);
    }

    fetchCharacters();
  }, []);

  return (
    <>
      <div className="flex flex-wrap gap-5 justify-center items-center my-5">
        {characters.map((character) => {
          return (
            <Link
              key={character.id}
              to={`/characters/$characterId`}
              params={{ characterId: character.id }}
            >
              <img src={character.spriteSPath} alt="" />
            </Link>
          );
        })}
      </div>
    </>
  );
}
