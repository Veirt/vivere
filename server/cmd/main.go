package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/gofiber/fiber/v3"
)

type Character struct {
	Id          string `json:"id"`
	SpriteSPath string `json:"spriteSPath"`
}

type MotionEntry struct {
	File string `json:"File"`
}

type ModelGroup struct {
	Target string   `json:"Target"`
	Name   string   `json:"Name"`
	Ids    []string `json:"Ids"`
}

type ModelFileReferences struct {
	Moc      string                   `json:"Moc"`
	Textures []string                 `json:"Textures"`
	Physics  string                   `json:"Physics"`
	Motions  map[string][]MotionEntry `json:"Motions"`
}

type ModelDetail struct {
	Version        int                 `json:"Version"`
	Name           string              `json:"Name"`
	FileReferences ModelFileReferences `json:"FileReferences"`
	Groups         []ModelGroup        `json:"Groups"`
}

type CharacterDetail struct {
	Id          string          `json:"id"`
	Sprites     CharacterSprite `json:"sprites"`
	Music       CharacterMusic  `json:"music"`
	ModelPath   string          `json:"modelPath"`
	ModelDetail ModelDetail     `json:"modelDetail"`
}

type CharacterSprite struct {
	SpriteSPath string `json:"spriteSPath"`
	SpriteMPath string `json:"spriteMPath"`
	SpriteLPath string `json:"spriteLPath"`
	SpriteWPath string `json:"spriteWPath"`
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

func resolveAudioURL(assetsDir string, fileName string) string {
	diskPath := filepath.Join(assetsDir, "AudioClip", fileName)
	if _, err := os.Stat(diskPath); err == nil {
		return "/api/assets/AudioClip/" + fileName
	}

	return ""
}

func GetCharacterMusic(assetsDir string, charId string) CharacterMusic {
	return CharacterMusic{
		FullJPMusicPath:  resolveAudioURL(assetsDir, charId+"_SONG_JP.wav"),
		FullENMusicPath:  resolveAudioURL(assetsDir, charId+"_SONG_US.wav"),
		ShortJPMusicPath: resolveAudioURL(assetsDir, charId+"_SONG_SHORT_JP.wav"),
		ShortENMusicPath: resolveAudioURL(assetsDir, charId+"_SONG_SHORT_US.wav"),
	}
}

type CharacterJSON struct {
	Characters []Character `json:"characters"`
}

type Response struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func LoadCharacters(assetsDir string) (CharacterJSON, error) {
	modelsDir := filepath.Join(assetsDir, "Models")
	entries, err := os.ReadDir(modelsDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return CharacterJSON{Characters: []Character{}}, nil
		}

		return CharacterJSON{}, fmt.Errorf("read models directory %q: %w", modelsDir, err)
	}

	characters := make([]Character, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		id := entry.Name()
		characters = append(characters, Character{
			Id:          id,
			SpriteSPath: GetCharacterSprites(id).SpriteSPath,
		})
	}

	return CharacterJSON{Characters: characters}, nil
}

func readModelDetail(modelFilePath string) (ModelDetail, error) {
	jsonFile, err := os.Open(modelFilePath)
	if err != nil {
		return ModelDetail{}, fmt.Errorf("open model detail %q: %w", modelFilePath, err)
	}
	defer func() {
		if closeErr := jsonFile.Close(); closeErr != nil {
			log.Printf("close model detail %q: %v", modelFilePath, closeErr)
		}
	}()

	payload, err := io.ReadAll(jsonFile)
	if err != nil {
		return ModelDetail{}, fmt.Errorf("read model detail %q: %w", modelFilePath, err)
	}

	var detail ModelDetail
	if err := json.Unmarshal(payload, &detail); err != nil {
		return ModelDetail{}, fmt.Errorf("unmarshal model detail %q: %w", modelFilePath, err)
	}

	return detail, nil
}

func NewApp(assetsDir string) (*fiber.App, error) {
	characters, err := LoadCharacters(assetsDir)
	if err != nil {
		return nil, err
	}

	knownIDs := make(map[string]struct{}, len(characters.Characters))
	for _, character := range characters.Characters {
		knownIDs[character.Id] = struct{}{}
	}

	app := fiber.New()
	app.Static("/api/assets", assetsDir)

	app.Get("/api/characters", func(c fiber.Ctx) error {
		return c.JSON(characters)
	})

	app.Get("/api/characters/:characterId", func(c fiber.Ctx) error {
		characterId := c.Params("characterId")
		if _, ok := knownIDs[characterId]; !ok {
			return c.Status(fiber.StatusNotFound).JSON(&Response{
				Success: false,
				Message: "Invalid id given",
			})
		}

		modelFilePath := filepath.Join(assetsDir, "Models", characterId, "model", "model.model3.json")
		detail, err := readModelDetail(modelFilePath)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return c.Status(fiber.StatusNotFound).JSON(&Response{
					Success: false,
					Message: "Model definition not found",
				})
			}

			return c.Status(fiber.StatusInternalServerError).JSON(&Response{
				Success: false,
				Message: "Error when reading model",
			})
		}

		return c.JSON(&CharacterDetail{
			Id:          characterId,
			Sprites:     GetCharacterSprites(characterId),
			Music:       GetCharacterMusic(assetsDir, characterId),
			ModelPath:   "/api/assets/Models/" + characterId + "/model/model.model3.json",
			ModelDetail: detail,
		})
	})

	return app, nil
}

func main() {
	assetsDir := os.Getenv("VIVERE_ASSETS_DIR")
	if assetsDir == "" {
		assetsDir = "./assets"
	}

	app, err := NewApp(assetsDir)
	if err != nil {
		log.Fatalf("initialize server: %v", err)
	}

	addr := os.Getenv("VIVERE_LISTEN_ADDR")
	if addr == "" {
		addr = ":3000"
	}

	if err := app.Listen(addr); err != nil {
		log.Fatalf("listen on %s: %v", addr, err)
	}
}
