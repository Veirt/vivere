import { Application } from "@pixi/app";
import { InteractionManager } from "@pixi/interaction";
import { Ticker, TickerPlugin } from "@pixi/ticker";
import { Live2DModel, MotionPriority } from "pixi-live2d-display/cubism4";
import * as PIXI from "pixi.js";
import { useEffect, useState } from "react";

// @ts-ignore
window.PIXI = PIXI;

// register InteractionManager to make Live2D models interactive
Live2DModel.registerTicker(Ticker);

PIXI.extensions.add(PIXI.InteractionManager);

function App() {
  useEffect(() => {
    const app = new Application({
      view: document.getElementById("canvas") as HTMLCanvasElement,
      autoStart: true,
    });

    Live2DModel.from("CHR_000085/model.model3.json").then((model) => {
      app.stage.addChild(model);
      model.anchor.set(0.5, 0.5);
      model.position.set(window.innerWidth / 4, window.innerHeight / 4);
      model.scale.set(0.1, 0.1);
      model.motion("EF_Idle.anim", 0, MotionPriority.IDLE);

      // model.on("click", (hitAreas) => {
      //   model.motion("Skill.anim", 0, MotionPriority.FORCE);
      // });
    });
  }, []);

  return (
    <>
      <canvas id="canvas" />;
    </>
  );
}

export default App;
