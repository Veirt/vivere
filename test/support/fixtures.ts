export const sampleCharacterListJson = `{
  "characters": [
    {
      "id": "CHR_000085",
      "spriteSPath": "/api/assets/Sprite/CHR_000085_00_s.png"
    }
  ]
}`;

export const sampleCharacterDetailJson = `{
  "id": "CHR_000085",
  "sprites": {
    "spriteSPath": "/api/assets/Sprite/CHR_000085_00_s.png",
    "spriteMPath": "/api/assets/Sprite/CHR_000085_00_m.png",
    "spriteLPath": "/api/assets/Sprite/CHR_000085_00_l.png",
    "spriteWPath": "/api/assets/Sprite/CHR_000085_00_w.png"
  },
  "music": {
    "fullJPMusicPath": "/api/assets/AudioClip/CHR_000085_SONG_JP.wav",
    "fullENMusicPath": "",
    "shortJPMusicPath": "/api/assets/AudioClip/CHR_000085_SONG_SHORT_JP.wav",
    "shortENMusicPath": ""
  },
  "modelPath": "/api/assets/Models/CHR_000085/model/model.model3.json",
  "modelDetail": {
    "Version": 3,
    "Name": "model",
    "FileReferences": {
      "Moc": "model.moc3",
      "Textures": ["textures/texture_00.png"],
      "Physics": "model.physics3.json",
      "Motions": {
        "Idle.anim": [{ "File": "motions/Idle.anim.motion3.json" }],
        "Face_Smile.anim": [{ "File": "motions/Face_Smile.anim.motion3.json" }],
        "Skill.anim": [
          { "File": "motions/Skill.anim.motion3.json" },
          { "File": "motions/Skill_Alt.anim.motion3.json" }
        ]
      }
    },
    "Groups": [
      {
        "Target": "Parameter",
        "Name": "EyeBlink",
        "Ids": ["ParamEyeLOpen", "ParamEyeOpenUpaL", "ParamEyeROpen", "ParamEyeOpenUpaR"]
      },
      {
        "Target": "Parameter",
        "Name": "LipSync",
        "Ids": ["ParamMouthOpenY"]
      }
    ]
  }
}`;
