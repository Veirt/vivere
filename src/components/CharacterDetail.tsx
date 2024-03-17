import {
  Cubism4InternalModel,
  InternalModel,
  Live2DModel,
  MotionPreloadStrategy,
  MotionPriority,
} from "pixi-live2d-display/cubism4";
import * as PIXI from "pixi.js";
import { useEffect, useRef, useState } from "react";

// @ts-ignore
window.PIXI = PIXI;

PIXI.extensions.add(PIXI.InteractionManager);

// register InteractionManager to make Live2D models interactive
Live2DModel.registerTicker(PIXI.Ticker);

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
  modelDetail: {
    Version: number;
    Name: string;
    FileReferences: {
      Moc: string;
      Textures: string[];
      Physics: string;
      Motions: {
        [key: string]: { File: string };
      };
    };
    Groups: {
      Target: string;
      Name: string;
      Ids: string;
    }[];
  };
};

export default function CharacterDetail({ id }: Props) {
  const [charDetail, setCharDetail] = useState<CharacterDetail | null>(null);

  const canvasRef = useRef<HTMLCanvasElement>(null);
  const [model, setModel] = useState<Live2DModel<InternalModel>>();

  function handleMotionClick(motion: string) {
    if (!model) return;
    model.motion(motion, 0, MotionPriority.FORCE);
  }

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
        backgroundAlpha: 0,
        width: 500,
      });

      const m = await Live2DModel.from(modelPath, {
        motionPreload: MotionPreloadStrategy.IDLE,
        idleMotionGroup: "Idle.anim",
        autoInteract: false,
      });

      // @ts-expect-error
      (m.internalModel as Cubism4InternalModel).breath = null;

      m.scale.set(0.12, 0.12);
      m.anchor.set(0.5, 0.5);
      m.position.set(
        canvasRef.current!.width / 2,
        canvasRef.current!.height / 2,
      );

      app.stage.addChild(m);
      setModel(m);
    }

    fetchCharacterDetail().then(async (characterDetail) => {
      console.log(characterDetail.modelPath);

      await renderLive2d(characterDetail.modelPath);
    });
  }, []);

  return (
    <>
      <main className="flex">
        {charDetail?.id}
        <canvas
          style={{ background: "url('/assets/Background_DungeonBattle.png')" }}
          className="rounded"
          ref={canvasRef}
          id="canvas"
        ></canvas>

        <div className="flex flex-col flex-wrap">
          {charDetail &&
            Object.keys(charDetail!.modelDetail.FileReferences.Motions).map(
              (key) => {
                return (
                  <button key={key} onClick={() => handleMotionClick(key)}>
                    {key}
                  </button>
                );
              },
            )}
        </div>
      </main>
    </>
  );
}
