import { Live2DModel } from "pixi-live2d-display/cubism4";
import * as PIXI from "pixi.js";
import { Ticker } from "pixi.js";
import { useEffect, useRef, useState } from "react";

// @ts-ignore
window.PIXI = PIXI;

PIXI.extensions.add(PIXI.InteractionManager);

// register InteractionManager to make Live2D models interactive
Live2DModel.registerTicker(Ticker);

type Props = {
  id: string;
};

type CharacterDetail = {
  id: string;
  sprites: {
    spriteSPath: string;
    spriteMPath: string;
    spriteLPath: string;
    spriteWPath: string;
  };
  music: {
    fullJPMusicPath: string;
    fullENMusicPath: string;
    shortJPMusicPath: string;
    shortENMusicPath: string;
  };
  modelPath: string;
};

export default function CharacterDetail({ id }: Props) {
  const [charDetail, setCharDetail] = useState<CharacterDetail | null>(null);

  const canvasRef = useRef<HTMLCanvasElement>(null);

  useEffect(() => {
    async function fetchCharacterDetail(): Promise<CharacterDetail> {
      return new Promise(async (resolve, reject) => {
        try {
          const res = await fetch(`/api/characters/${id}`);
          const data = await res.json();
          setCharDetail(data);

          return resolve(data);
        } catch (err) {
          return reject(err);
        }
      });
    }

    async function renderLive2d(modelPath: string) {
      const app = new PIXI.Application({
        view: document.getElementById("canvas") as HTMLCanvasElement,
      });
      app.renderer.backgroundAlpha = 0;

      const model = await Live2DModel.from(modelPath);
      model.motion("EF_Idle.anim");

      model.x = 100;
      model.y = 100;
      model.scale.set(0.1, 0.1);
      model.anchor.set(0.5, 0.5);
      model.position.set(
        canvasRef.current!.width / 2,
        canvasRef.current!.height / 2,
      );

      app.stage.addChild(model);
    }

    fetchCharacterDetail().then(async (characterDetail) => {
      await renderLive2d(characterDetail.modelPath);
    });
  }, []);

  return (
    <>
      {charDetail?.id}
      <canvas
        style={{ background: "url('/assets/Background_DungeonBattle.png')" }}
        className="rounded"
        ref={canvasRef}
        id="canvas"
      ></canvas>
    </>
  );
}
