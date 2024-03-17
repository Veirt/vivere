package main

import (
	"encoding/json"
	"io"
	"log"
	"os"

	"github.com/gofiber/fiber/v3"
)

type Character struct {
	Id          string `json:"id"`
	SpriteSPath string `json:"spriteSPath"` // Small
}

type ModelDetail struct {
	Version        int    `json:"Version"`
	Name           string `json:"Name"`
	FileReferences struct {
		Moc      string   `json:"Moc"`
		Textures []string `json:"Textures"`
		Physics  string   `json:"Physics"`
		Motions  map[string][]struct {
			File string `json:"File"`
		} `json:"Motions"`
		// Expressions? not sure
	} `json:"FileReferences"`
	Groups []struct {
		Target string   `json:"Target"`
		Name   string   `json:"Name"`
		Ids    []string `json:"Ids"`
	} `json:"Groups"`
}

type CharacterDetail struct {
	Id          string          `json:"id"`
	Sprites     CharacterSprite `json:"sprites"`
	Music       CharacterMusic  `json:"music"`
	ModelPath   string          `json:"modelPath"`
	ModelDetail ModelDetail     `json:"modelDetail"`
}

type CharacterSprite struct {
	SpriteSPath string `json:"spriteSPath"` // Small
	SpriteMPath string `json:"spriteMPath"` // Medium?
	SpriteLPath string `json:"spriteLPath"` // Large?
	SpriteWPath string `json:"spriteWPath"` // most definitely wide
}

func GetCharacterSprites(charId string) CharacterSprite {
	return CharacterSprite{
		SpriteSPath: "/api/assets/Sprite/" + charId + "_00_s.png",
		SpriteMPath: "/api/assets/Sprite/" + charId + "_00_m.png",
		SpriteLPath: "/api/assets/Sprite/" + charId + "_00_l.png",
		SpriteWPath: "/api/assets/Sprite/" + charId + "_00_w.png",
	}
}

type CharacterMusic struct {
	FullJPMusicPath  string `json:"fullJPMusicPath"`
	FullENMusicPath  string `json:"fullENMusicPath"`
	ShortJPMusicPath string `json:"shortJPMusicPath"`
	ShortENMusicPath string `json:"shortENMusicPath"`
}

func GetCharacterMusic(charId string) CharacterMusic {
	return CharacterMusic{
		FullJPMusicPath:  "/api/assets/AudioClip/" + charId + "_SONG_JP.wav",
		FullENMusicPath:  "/api/assets/AudioClip/" + charId + "_SONG_US.wav",
		ShortJPMusicPath: "/api/assets/AudioClip/" + charId + "_SONG_SHORT_JP.wav",
		ShortENMusicPath: "/api/assets/AudioClip/" + charId + "_SONG_SHORT_US.wav",
	}
}

type CharacterJSON struct {
	Characters []Character `json:"characters"`
}

type Response struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func LoadCharacters() CharacterJSON {
	files, err := os.ReadDir("assets/Models")
	if err != nil {
		log.Fatal(err)
	}

	cl := CharacterJSON{Characters: []Character{}}
	for _, file := range files {

		id := file.Name()

		ch := Character{
			Id:          id,
			SpriteSPath: GetCharacterSprites(id).SpriteSPath,
		}

		cl.Characters = append(cl.Characters, ch)
	}

	return cl
}

func main() {
	app := fiber.New()
	app.Static("/api/assets", "./assets")

	cl := LoadCharacters()

	app.Get("/api/characters", func(c fiber.Ctx) error {
		return c.JSON(cl)
	})

	app.Get("/api/characters/:characterId", func(c fiber.Ctx) error {
		characterId := c.Params("characterId")
		for _, char := range cl.Characters {
			if char.Id == characterId {
				jsonFile, err := os.Open("./assets/Models/" + char.Id + "/model/model.model3.json")
				if err != nil {
					return c.Status(400).JSON(&Response{
						Success: false,
						Message: "Error when reading model",
					})
				}

				var md ModelDetail
				b, _ := io.ReadAll(jsonFile)
				json.Unmarshal(b, &md)

				return c.JSON(&CharacterDetail{
					Id:          char.Id,
					Sprites:     GetCharacterSprites(char.Id),
					Music:       GetCharacterMusic(char.Id),
					ModelPath:   "/api/assets/Models/" + char.Id + "/model/model.model3.json",
					ModelDetail: md,
				})
			}
		}

		return c.Status(400).JSON(&Response{
			Success: false,
			Message: "Invalid id given",
		})
	})

	app.Listen(":3000")
}
