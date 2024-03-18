import {
	Cubism4InternalModel,
	InternalModel,
	Live2DModel,
	MotionPriority,
} from "pixi-live2d-display/cubism4";
import * as PIXI from "pixi.js";
import { WheelEvent, useEffect, useRef, useState } from "react";

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
			});

			const m = await Live2DModel.from(modelPath, {
				idleMotionGroup: "Idle.anim",
				autoInteract: false,
				// autoUpdate: false, // TODO: remove later
			});

			// @ts-expect-error
			(m.internalModel as Cubism4InternalModel).breath = null;

			m.scale.set(0.135, 0.12);
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
		<div
			style={{
				background: "url('/api/assets/Background_DungeonBattle.png')",
			}}
		>
			<main className="mx-5 min-h-[95vh]">
				<h1 className="my-5 text-xl font-bold">Live2D Model</h1>

				<div className="flex">
					<canvas
						onWheelCapture={handleZoomModel}
						className={`rounded-md border-3 border-primary-200 ${
							!model ? "invisible" : ""
						}`}
						width={800}
						ref={canvasRef}
						id="canvas"
					></canvas>

					<div className="flex flex-col">
						<h1 className="mx-3 mb-5 text-xl font-bold">Character Name</h1>
						<div className="flex flex-row flex-wrap gap-2 mx-3 w-1/2">
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
		</div>
	);
}
