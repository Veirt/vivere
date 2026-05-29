package webapi

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/veirt/vivere/internal/masterdata"
)

func (a API) enrichCharacterAudio(characters []masterdata.Character) {
	if a.exportsRoot == "" {
		return
	}
	for characterIndex := range characters {
		character := &characters[characterIndex]
		for storyIndex := range character.Stories {
			story := &character.Stories[storyIndex]
			story.AudioJP = a.memoryAudioPaths("JP", character.ID, story.EpisodeID)
			story.AudioUS = a.memoryAudioPaths("US", character.ID, story.EpisodeID)
		}
		for voiceIndex := range character.Voices {
			voice := &character.Voices[voiceIndex]
			voice.AudioJP = a.commonVoicePath("JP", character.ID, voice.SortOrder)
			voice.AudioUS = a.commonVoicePath("US", character.ID, voice.SortOrder)
		}
	}
}

func (a API) memoryAudioPaths(region string, characterID int64, episodeID int64) []string {
	pattern := filepath.Join(
		a.exportsRoot,
		"Assets",
		"AddressableConvertAssets",
		"Voice",
		region,
		"Character",
		fmt.Sprintf("CHR_%06d", characterID),
		"Memory",
		fmt.Sprintf("MEM_%06d", episodeID),
		fmt.Sprintf("MV_%s_%06d_%02d_*.wav", region, characterID, episodeID),
	)
	matches, err := filepath.Glob(pattern)
	if err != nil || len(matches) == 0 {
		return []string{}
	}
	sort.Strings(matches)
	paths := make([]string, 0, len(matches))
	for _, match := range matches {
		paths = append(paths, exportAssetPath(match, a.exportsRoot))
	}
	return paths
}

func (a API) commonVoicePath(region string, characterID int64, sortOrder int64) string {
	if sortOrder <= 0 {
		return ""
	}
	path := filepath.Join(
		a.exportsRoot,
		"Assets",
		"AddressableConvertAssets",
		"Voice",
		region,
		"Character",
		fmt.Sprintf("CHR_%06d", characterID),
		"Common",
		fmt.Sprintf("CV_%s_%06d_Mypage_%03d.wav", region, characterID, sortOrder),
	)
	if !regularFileExists(path) {
		return ""
	}
	return exportAssetPath(path, a.exportsRoot)
}

func exportAssetPath(path string, exportsRoot string) string {
	rel, err := filepath.Rel(exportsRoot, path)
	if err != nil {
		return ""
	}
	return "/assets/" + filepath.ToSlash(rel)
}

func regularFileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
