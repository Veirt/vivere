package webapi

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"

	"github.com/veirt/vivere/internal/masterdata"
)

type dirFingerprint struct {
	exists          bool
	modTimeUnixNano int64
}

type commonCacheKey struct {
	region      string
	characterID int64
}

type commonCacheEntry struct {
	fp    dirFingerprint
	files map[string]struct{}
}

type memoryCacheKey struct {
	region      string
	characterID int64
	episodeID   int64
}

type memoryCacheEntry struct {
	fp    dirFingerprint
	paths []string
}

type characterAudioCache struct {
	mu     sync.RWMutex
	common map[commonCacheKey]commonCacheEntry
	memory map[memoryCacheKey]memoryCacheEntry
}

func newCharacterAudioCache() *characterAudioCache {
	return &characterAudioCache{
		common: make(map[commonCacheKey]commonCacheEntry),
		memory: make(map[memoryCacheKey]memoryCacheEntry),
	}
}

func (a API) enrichCharacterAudio(characters []masterdata.Character) {
	if a.exportsRoot == "" {
		return
	}
	for i := range characters {
		a.enrichOneCharacterAudio(&characters[i])
	}
}

func (a API) enrichOneCharacterAudio(character *masterdata.Character) {
	if a.exportsRoot == "" || character == nil {
		return
	}
	for storyIndex := range character.Stories {
		story := &character.Stories[storyIndex]
		story.AudioJP = a.memoryAudioPaths("JP", character.ID, story.EpisodeID)
		story.AudioUS = a.memoryAudioPaths("US", character.ID, story.EpisodeID)
	}
	if len(character.Voices) == 0 {
		return
	}
	jpDir, jpFiles := a.commonVoiceFiles("JP", character.ID)
	usDir, usFiles := a.commonVoiceFiles("US", character.ID)
	for voiceIndex := range character.Voices {
		voice := &character.Voices[voiceIndex]
		voice.AudioJP = a.lookupCommonVoice(jpDir, jpFiles, "JP", character.ID, voice.SortOrder)
		voice.AudioUS = a.lookupCommonVoice(usDir, usFiles, "US", character.ID, voice.SortOrder)
	}
}

func (a API) memoryAudioPaths(region string, characterID int64, episodeID int64) []string {
	dir := filepath.Join(
		a.exportsRoot,
		"Assets",
		"AddressableConvertAssets",
		"Voice",
		region,
		"Character",
		fmt.Sprintf("CHR_%06d", characterID),
		"Memory",
		fmt.Sprintf("MEM_%06d", episodeID),
	)
	fp := statDir(dir)
	key := memoryCacheKey{region: region, characterID: characterID, episodeID: episodeID}

	if a.audioCache != nil {
		a.audioCache.mu.RLock()
		entry, ok := a.audioCache.memory[key]
		a.audioCache.mu.RUnlock()
		if ok && entry.fp == fp {
			return cloneAudioPaths(entry.paths)
		}
	}

	paths := []string{}
	if fp.exists {
		prefix := fmt.Sprintf("MV_%s_%06d_%02d_", region, characterID, episodeID)
		entries, err := os.ReadDir(dir)
		if err == nil {
			matched := make([]string, 0, len(entries))
			for _, entry := range entries {
				name := entry.Name()
				if !strings.HasPrefix(name, prefix) || !strings.HasSuffix(name, ".wav") {
					continue
				}
				fullPath := filepath.Join(dir, name)
				if !dirEntryIsRegularFile(entry, fullPath) {
					continue
				}
				matched = append(matched, fullPath)
			}
			slices.Sort(matched)
			paths = make([]string, 0, len(matched))
			for _, match := range matched {
				paths = append(paths, exportAssetPath(match, a.exportsRoot))
			}
		}
	}

	if a.audioCache != nil {
		a.audioCache.mu.Lock()
		a.audioCache.memory[key] = memoryCacheEntry{fp: fp, paths: paths}
		a.audioCache.mu.Unlock()
	}
	return cloneAudioPaths(paths)
}

func (a API) commonVoiceFiles(region string, characterID int64) (string, map[string]struct{}) {
	dir := filepath.Join(
		a.exportsRoot,
		"Assets",
		"AddressableConvertAssets",
		"Voice",
		region,
		"Character",
		fmt.Sprintf("CHR_%06d", characterID),
		"Common",
	)
	fp := statDir(dir)
	key := commonCacheKey{region: region, characterID: characterID}

	if a.audioCache != nil {
		a.audioCache.mu.RLock()
		entry, ok := a.audioCache.common[key]
		a.audioCache.mu.RUnlock()
		if ok && entry.fp == fp {
			return dir, entry.files
		}
	}

	var files map[string]struct{}
	if fp.exists {
		entries, err := os.ReadDir(dir)
		if err == nil {
			files = make(map[string]struct{}, len(entries))
			for _, entry := range entries {
				name := entry.Name()
				if !strings.HasSuffix(name, ".wav") {
					continue
				}
				if !dirEntryIsRegularFile(entry, filepath.Join(dir, name)) {
					continue
				}
				files[name] = struct{}{}
			}
		}
	}

	if a.audioCache != nil {
		a.audioCache.mu.Lock()
		a.audioCache.common[key] = commonCacheEntry{fp: fp, files: files}
		a.audioCache.mu.Unlock()
	}
	return dir, files
}

func (a API) lookupCommonVoice(dir string, files map[string]struct{}, region string, characterID int64, sortOrder int64) string {
	if sortOrder <= 0 || len(files) == 0 {
		return ""
	}
	filename := fmt.Sprintf("CV_%s_%06d_Mypage_%03d.wav", region, characterID, sortOrder)
	if _, ok := files[filename]; !ok {
		return ""
	}
	return exportAssetPath(filepath.Join(dir, filename), a.exportsRoot)
}

func exportAssetPath(path string, exportsRoot string) string {
	rel, err := filepath.Rel(exportsRoot, path)
	if err != nil {
		return ""
	}
	return "/assets/" + filepath.ToSlash(rel)
}

func statDir(path string) dirFingerprint {
	info, err := os.Stat(path)
	if err != nil || !info.IsDir() {
		return dirFingerprint{}
	}
	return dirFingerprint{
		exists:          true,
		modTimeUnixNano: info.ModTime().UnixNano(),
	}
}

func dirEntryIsRegularFile(entry os.DirEntry, fullPath string) bool {
	if entry.IsDir() {
		return false
	}
	if entry.Type()&os.ModeSymlink != 0 {
		return regularFileExists(fullPath)
	}
	return true
}

func regularFileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func cloneAudioPaths(paths []string) []string {
	if len(paths) == 0 {
		return []string{}
	}
	return slices.Clone(paths)
}
