import { fireEvent, render, screen, within } from "@testing-library/react";
import { describe, expect, it } from "vite-plus/test";

import type { Character } from "#/lib/characters.ts";
import { CharacterDetail } from "./characters.$characterId.tsx";

describe("CharacterDetail", () => {
  it("renders rich profile, combat, lament, memories, and voicelines safely", () => {
    render(<CharacterDetail character={character()} />);

    expect(screen.getByRole("heading", { name: "Iris" })).toBeTruthy();
    expect(screen.getByText("The Bloom")).toBeTruthy();
    expect(screen.getAllByText("Crimson").length).toBeGreaterThan(0);
    expect(screen.getAllByText("Sniper").length).toBeGreaterThan(0);
    expect(screen.getByText("03/07")).toBeTruthy();
    expect(screen.getByText("162 cm")).toBeTruthy();
    expect(screen.getByText("Kana Voice")).toBeTruthy();
    expect(screen.getByText("Ava Voice")).toBeTruthy();
    selectTab("Skills");
    expect(screen.getByText("Bloom Strike")).toBeTruthy();
    expect(screen.getByText("A bright attack.")).toBeTruthy();
    expect(screen.getByText("Quiet Resolve")).toBeTruthy();

    selectTab("Unique Weapon");
    expect(screen.getByText("Iris Tome")).toBeTruthy();
    expect(screen.getByText("LR weapon bonus.")).toBeTruthy();

    selectTab("Stats");
    expect(screen.getByText("Normal 1001")).toBeTruthy();
    expect(screen.getByText("Active 1002")).toBeTruthy();
    expect(screen.getByText("Muscle")).toBeTruthy();
    expect(screen.getByText("90")).toBeTruthy();

    selectTab("Lament");
    const lament = screen.getByRole("heading", { name: "Lament" }).closest('[data-slot="card"]');
    expect(lament).not.toBeNull();
    expect(within(lament as HTMLElement).getByText("JP: JP Lament")).toBeTruthy();
    expect(
      within(lament as HTMLElement)
        .getByRole("link", { name: "YouTube US" })
        .getAttribute("href"),
    ).toBe("https://youtu.be/us");
    expect(within(lament as HTMLElement).getByText("US lyrics")).toBeTruthy();

    selectTab("Memories");
    expect(screen.getByRole("heading", { name: "Memories" })).toBeTruthy();
    expect(screen.getByText("First memory")).toBeTruthy();
    expect(screen.getByText("The story text.")).toBeTruthy();
    expect(screen.getByLabelText("JP segment 1")).toBeTruthy();
    selectTab("Voicelines");
    expect(screen.getByRole("heading", { name: "Voicelines" })).toBeTruthy();
    expect(screen.getByText("Greeting")).toBeTruthy();
    expect(screen.getByText("Hello.")).toBeTruthy();
    expect(screen.getByLabelText("JP voice")).toBeTruthy();
  });
});

function selectTab(name: string) {
  fireEvent.mouseDown(screen.getByRole("tab", { name }), { button: 0, ctrlKey: false });
}

function character(): Character {
  return {
    id: 101,
    memo: "test character",
    name: "Iris",
    title: "The Bloom",
    nameKey: "character.name.101",
    name2Key: "character.title.101",
    rarityFlags: 4,
    elementType: 2,
    jobFlags: 2,
    characterType: 1,
    startTimeFixJST: "2024-01-01 00:00:00",
    endTimeFixJST: "2100-12-31 23:59:59",
    iconPath: "/assets/icon.png",
    profile: {
      birthday: 307,
      height: 162,
      weight: 45,
      bloodType: 4,
      description: "A gentle singer.",
      voiceJp: "Kana Voice",
      voiceUs: "Ava Voice",
      vocalJp: "Singer JP",
      vocalUs: "Singer US",
      lamentJp: "JP Lament",
      lamentUs: "US Lament",
      lyricsJp: "JP lyrics",
      lyricsUs: "US lyrics",
      gachaMessage: "I will join you.",
      movieJpUrls: { jaJP: "https://youtu.be/jp" },
      movieUsUrls: { enUS: "https://youtu.be/us" },
    },
    stats: {
      speed: 2733,
      normalSkillId: 1001,
      activeSkillIds: [1002],
      passiveSkillIds: [1003],
      baseParameterCoefficient: { Muscle: 90, Energy: 80 },
      baseParameterGrossCoefficient: 341,
      initialBattleParameter: { Speed: 2733, Defense: 12 },
    },
    skills: {
      normal: {
        id: 1001,
        type: "active",
        name: "Normal Attack",
        nameKey: "skill.normal",
        icon: "/assets/skill-normal.png",
        conditions: "True",
        initCooltime: 0,
        maxCooltime: 0,
        levels: [
          {
            orderNumber: 1,
            characterLevel: 1,
            equipmentRarityFlags: 0,
            description: "Normal attack.",
            descriptionKey: "skill.normal.description",
            subSkillIds: [],
          },
        ],
      },
      active: [
        {
          id: 1002,
          type: "active",
          name: "Bloom Strike",
          nameKey: "skill.active",
          icon: "/assets/skill-active.png",
          conditions: "True",
          initCooltime: 1,
          maxCooltime: 4,
          levels: [
            {
              orderNumber: 1,
              characterLevel: 11,
              equipmentRarityFlags: 0,
              description: "A bright attack.",
              descriptionKey: "skill.active.description",
              subSkillIds: [],
            },
          ],
        },
      ],
      passive: [
        {
          id: 1003,
          type: "passive",
          name: "Quiet Resolve",
          nameKey: "skill.passive",
          icon: "/assets/skill-passive.png",
          conditions: "",
          initCooltime: 0,
          maxCooltime: 0,
          levels: [
            {
              orderNumber: 1,
              characterLevel: 41,
              equipmentRarityFlags: 0,
              description: "Raises defense.",
              descriptionKey: "skill.passive.description",
              subSkillIds: [],
            },
          ],
        },
      ],
    },
    uniqueWeapon: {
      equipmentId: 9001,
      name: "Iris Tome",
      nameKey: "equipment.unique",
      iconId: 777,
      icon: "/assets/equipment.png",
      jobFlags: 2,
      rarityFlags: 128,
      level: 180,
      slotType: 1,
      compositeId: 33,
      exclusiveEffectId: 44,
      skillDescriptionId: 55,
      battleParameter: { Value: 500 },
      bonusParameters: [{ category: "base", type: 1, changeType: 2, value: 1200 }],
      descriptions: { ssr: "SSR weapon bonus.", ur: "UR weapon bonus.", lr: "LR weapon bonus." },
    },
    assets: {
      iconSmall: "/assets/icon-small.png",
      iconMedium: "/assets/icon-medium.png",
      iconLarge: "/assets/icon-large.png",
      iconWide: "/assets/icon-wide.png",
      iconEmotion: "/assets/icon-emotion.png",
      modelTexture: "/assets/model.png",
      effectTexture: "/assets/effect.png",
      songJp: "/assets/song-jp.wav",
      songUs: "/assets/song-us.wav",
      songShortJp: "/assets/song-short-jp.wav",
      songShortUs: "/assets/song-short-us.wav",
      elementIcon: "/assets/icon-element.png",
      jobIcon: "/assets/icon-job.png",
    },
    stories: [
      {
        id: 1,
        episodeId: 1,
        level: 0,
        rarityFlags: 1,
        title: "First memory",
        text: "The story text.",
        titleKey: "story.title.101",
        textKey: "story.text.101",
        rewardItems: [],
        audioJp: ["/assets/memory-jp-1.wav"],
        audioUs: [],
      },
    ],
    voices: [
      {
        id: 2,
        category: 3,
        unlockCondition: 0,
        unlockQuestId: 0,
        sortOrder: 1,
        subtitle: "Hello.",
        subtitleKey: "voice.subtitle.101",
        buttonText: "Greeting",
        buttonTextKey: "voice.button.101",
        audioJp: "/assets/voice-jp.wav",
        audioUs: "",
      },
    ],
  };
}
