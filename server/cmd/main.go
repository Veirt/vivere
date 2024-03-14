package main

import (
	"github.com/gofiber/fiber/v3"
	"log"
	"os"
)

type Character struct {
	Id         string `json:"id"`
	SpritePath string `json:"spritePath"`
	ModelPath  string `json:"modelPath"`
}

type CharacterJSON struct {
	Characters []Character `json:"characters"`
}

func main() {
	app := fiber.New()
	app.Static("/assets", "./assets")

	app.Get("/api/characters", func(c fiber.Ctx) error {
		files, err := os.ReadDir("assets/Models")
		if err != nil {
			log.Fatal(err)
		}

		cl := &CharacterJSON{Characters: []Character{}}
		for _, file := range files {

			id := file.Name()

			ch := Character{
				Id:         id,
				ModelPath:  "/assets/Models/" + id + "/model/model.model3.json",
				SpritePath: "/assets/Sprite/" + id + "_00_s.png",
			}

			cl.Characters = append(cl.Characters, ch)
		}

		return c.JSON(cl)
	})

	app.Listen(":3000")
}
