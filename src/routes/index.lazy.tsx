import { Ticker } from "@pixi/ticker";
import { createLazyFileRoute } from "@tanstack/react-router";
import { Live2DModel } from "pixi-live2d-display/cubism4";
import * as PIXI from "pixi.js";
import { useState, useEffect } from "react";

// @ts-ignore
window.PIXI = PIXI;

// register InteractionManager to make Live2D models interactive
Live2DModel.registerTicker(Ticker);

PIXI.extensions.add(PIXI.InteractionManager);

type Character = {
  id: string;
  spritePath: string;
  modelPath: string;
};

function Index() {
  const [characters, setCharacters] = useState<Character[]>([]);
  useEffect(() => {
    async function fetchCharacters() {
      const res = await fetch("/api/characters");
      const data = await res.json();

      console.log(data);
      setCharacters(data.characters);
    }

    fetchCharacters();
  }, []);

  return (
    <>
      <div className="flex flex-wrap gap-5 justify-center items-center my-5">
        {characters.map((character) => {
          return <img src={character.spritePath} alt="" />;
        })}
      </div>
    </>
  );
}

export const Route = createLazyFileRoute("/")({
  component: Index,
});
