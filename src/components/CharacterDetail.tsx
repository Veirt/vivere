import {
  Cubism4InternalModel,
  InternalModel,
  Live2DModel,
  MotionPriority,
} from "pixi-live2d-display/cubism4";
import * as PIXI from "pixi.js";
import { WheelEvent, useEffect, useRef, useState } from "react";

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

  const audioRef = useRef<HTMLAudioElement>(null);
  audioRef.current?.play();
  if (audioRef.current) {
    audioRef.current.loop = true;
  }

  function handleMotionClick(motion: string) {
    if (!model) return;
    model.motion(motion, 0, MotionPriority.FORCE);
  }

  function handleZoomModel(e: WheelEvent<HTMLCanvasElement>) {
    if (!model) return;

    let x = model.scale._x;
    let y = model.scale._y;

    if (e.deltaY < 0) {
      x += 0.01;
      y += 0.01;
    } else {
      if (x <= 0.07 || y <= 0.05) return;
      x -= 0.01;
      y -= 0.01;
    }

    model.scale.set(x, y);
  }

  // https://stackoverflow.com/questions/57358640/cancel-wheel-event-with-e-preventdefault-in-react-event-bubbling
  const preventDefault = (e: Event) => e.preventDefault();
  useEffect(() => {
    canvasRef.current?.addEventListener("wheel", preventDefault);
    return () => {
      canvasRef.current?.removeEventListener("wheel", preventDefault);
    };
  }, []);

  useEffect(() => {
    async function fetchCharacterDetail(): Promise<CharacterDetail> {
      const res = await fetch(`/api/characters/${id}`);
      const data = await res.json();
      setCharDetail(data);

      return data;
    }

    async function renderLive2d(modelPath: string) {
      const app = new PIXI.Application({
        view: document.getElementById("canvas") as HTMLCanvasElement,
        backgroundAlpha: 0,
      });

      const m = await Live2DModel.from(modelPath, {
        idleMotionGroup: "Idle.anim",
        autoInteract: false,
        // autoUpdate: false, // TODO: remove later
      });

      // @ts-expect-error: Somehow, breathing motion is slightly off, will try to find a fix later
      (m.internalModel as Cubism4InternalModel).breath = null;

      m.scale.set(0.1, 0.1);
      m.anchor.set(0.5, 0.5);
      m.position.set(
        canvasRef.current!.width / 2,
        canvasRef.current!.height / 2,
      );

      app.stage.addChild(m);
      setModel(m);
    }

    fetchCharacterDetail().then(async (characterDetail) => {
      await renderLive2d(characterDetail.modelPath);
    });
  }, []);

  return (
    <>
      <main className="mx-5 min-h-[95vh]">
        <h1 className="my-5 text-xl font-bold">Live2D Model</h1>

        <div className="flex flex-col items-center sm:justify-start sm:items-start md:flex-row">
          <canvas
            onWheelCapture={handleZoomModel}
            className={`w-[90%]  md:w-1/2 rounded-md border-3 border-primary-200 ${
              !model ? "invisible" : ""
            }`}
            ref={canvasRef}
            id="canvas"
          ></canvas>

          <div className="flex flex-col w-[90%] md:w-1/2 ">
            <h1 className="mx-3 mb-5 text-xl font-bold">Character Name</h1>
            <div className="flex flex-row flex-wrap gap-2 mx-3">
              {charDetail &&
                Object.keys(charDetail.modelDetail.FileReferences.Motions).map(
                  (key) => {
                    return (
                      <button
                        key={key}
                        className="p-1 bg-black rounded text-primary-100"
                        onClick={() => handleMotionClick(key)}
                      >
                        {key}
                      </button>
                    );
                  },
                )}
            </div>
            <h1 className="my-5 mx-3 text-xl font-bold">Sprites</h1>
            <div className="flex flex-wrap gap-5 mx-3">
              {charDetail &&
                Object.values(charDetail.sprites).map((sprite) => (
                  <img
                    className="object-contain p-2 w-36 bg-white border border-primary-200"
                    key={sprite}
                    src={sprite}
                    alt=""
                  />
                ))}
            </div>
          </div>
        </div>
      </main>

      <audio
        ref={audioRef}
        src={
          charDetail?.music.shortJPMusicPath ||
          charDetail?.music.fullJPMusicPath
        }
      ></audio>
    </>
  );
}
