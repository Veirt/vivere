package masterdata

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"sync"

	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/singleflight"
)

type Character struct {
	ID              int64            `json:"id"`
	Memo            string           `json:"memo"`
	Name            string           `json:"name"`
	Title           string           `json:"title"`
	NameKey         string           `json:"nameKey"`
	Name2Key        string           `json:"name2Key"`
	RarityFlags     int64            `json:"rarityFlags"`
	ElementType     int64            `json:"elementType"`
	JobFlags        int64            `json:"jobFlags"`
	CharacterType   int64            `json:"characterType"`
	StartTimeFixJST string           `json:"startTimeFixJST"`
	EndTimeFixJST   string           `json:"endTimeFixJST"`
	Profile         Profile          `json:"profile"`
	Stats           CharacterStats   `json:"stats"`
	Assets          CharacterAssets  `json:"assets"`
	Skills          CharacterSkills  `json:"skills"`
	UniqueWeapon    UniqueWeapon     `json:"uniqueWeapon"`
	Stories         []CharacterStory `json:"stories"`
	Voices          []CharacterVoice `json:"voices"`
	IconPath        string           `json:"iconPath"`
}

type Profile struct {
	Birthday     int64             `json:"birthday"`
	Height       int64             `json:"height"`
	Weight       int64             `json:"weight"`
	BloodType    int64             `json:"bloodType"`
	Description  string            `json:"description"`
	VoiceJP      string            `json:"voiceJp"`
	VoiceUS      string            `json:"voiceUs"`
	VocalJP      string            `json:"vocalJp"`
	VocalUS      string            `json:"vocalUs"`
	LamentJP     string            `json:"lamentJp"`
	LamentUS     string            `json:"lamentUs"`
	LyricsJP     string            `json:"lyricsJp"`
	LyricsUS     string            `json:"lyricsUs"`
	GachaMessage string            `json:"gachaMessage"`
	MovieJPURLs  map[string]string `json:"movieJpUrls"`
	MovieUSURLs  map[string]string `json:"movieUsUrls"`
}

type CharacterStats struct {
	Speed                         int64            `json:"speed"`
	NormalSkillID                 int64            `json:"normalSkillId"`
	ActiveSkillIDs                []int64          `json:"activeSkillIds"`
	PassiveSkillIDs               []int64          `json:"passiveSkillIds"`
	BaseParameterCoefficient      map[string]int64 `json:"baseParameterCoefficient"`
	BaseParameterGrossCoefficient int64            `json:"baseParameterGrossCoefficient"`
	InitialBattleParameter        map[string]int64 `json:"initialBattleParameter"`
}

type CharacterAssets struct {
	IconSmall     string `json:"iconSmall"`
	IconMedium    string `json:"iconMedium"`
	IconLarge     string `json:"iconLarge"`
	IconWide      string `json:"iconWide"`
	IconEmotion   string `json:"iconEmotion"`
	ElementIcon   string `json:"elementIcon"`
	JobIcon       string `json:"jobIcon"`
	ModelTexture  string `json:"modelTexture"`
	EffectTexture string `json:"effectTexture"`
	SongJP        string `json:"songJp"`
	SongUS        string `json:"songUs"`
	SongShortJP   string `json:"songShortJp"`
	SongShortUS   string `json:"songShortUs"`
}

type CharacterStory struct {
	ID          int64        `json:"id"`
	EpisodeID   int64        `json:"episodeId"`
	Level       int64        `json:"level"`
	RarityFlags int64        `json:"rarityFlags"`
	Title       string       `json:"title"`
	Text        string       `json:"text"`
	TitleKey    string       `json:"titleKey"`
	TextKey     string       `json:"textKey"`
	RewardItems []RewardItem `json:"rewardItems"`
	AudioJP     []string     `json:"audioJp"`
	AudioUS     []string     `json:"audioUs"`
}

type CharacterVoice struct {
	ID              int64  `json:"id"`
	Category        int64  `json:"category"`
	UnlockCondition int64  `json:"unlockCondition"`
	UnlockQuestID   int64  `json:"unlockQuestId"`
	SortOrder       int64  `json:"sortOrder"`
	Subtitle        string `json:"subtitle"`
	SubtitleKey     string `json:"subtitleKey"`
	ButtonText      string `json:"buttonText"`
	ButtonTextKey   string `json:"buttonTextKey"`
	AudioJP         string `json:"audioJp"`
	AudioUS         string `json:"audioUs"`
}

type RewardItem struct {
	ItemID    int64 `json:"itemId"`
	ItemType  int64 `json:"itemType"`
	ItemCount int64 `json:"itemCount"`
}

var baseBookNames = [9]string{
	"CharacterMB",
	"CharacterProfileMB",
	"CharacterStoryMB",
	"CharacterDetailVoiceMB",
	"ActiveSkillMB",
	"PassiveSkillMB",
	"EquipmentMB",
	"EquipmentExclusiveEffectMB",
	"EquipmentExclusiveSkillDescriptionMB",
}

type fileFingerprint struct {
	exists          bool
	size            int64
	modTimeUnixNano int64
}

type baseDatasetFingerprint [len(baseBookNames)]fileFingerprint

type baseDataset struct {
	fingerprint           baseDatasetFingerprint
	characterRows         []map[string]any
	profiles              map[int64]map[string]any
	storyRows             []map[string]any
	voiceRows             []map[string]any
	activeSkills          map[int64]map[string]any
	passiveSkills         map[int64]map[string]any
	uniqueWeapons         map[int64]map[string]any
	exclusiveEffects      map[int64]map[string]any
	exclusiveDescriptions map[int64]map[string]any
}

type localizedCatalog struct {
	baseFingerprint baseDatasetFingerprint
	textFingerprint fileFingerprint
	characters      []Character
	indexByID       map[int64]int
}

type characterCache struct {
	mu           sync.RWMutex
	baseGroup    singleflight.Group
	catalogGroup singleflight.Group
	base         *baseDataset
	catalogs     map[string]*localizedCatalog
}

func newCharacterCache() *characterCache {
	return &characterCache{
		catalogs: make(map[string]*localizedCatalog),
	}
}

func (l Loader) LoadTextResources(bookName string) (map[string]string, error) {
	rows, err := l.LoadRows(bookName)
	if err != nil {
		return nil, err
	}
	texts := make(map[string]string, len(rows))
	for _, row := range rows {
		key := asString(row["StringKey"])
		if key == "" {
			continue
		}
		texts[key] = asString(row["Text"])
	}
	return texts, nil
}

func (l Loader) LoadCharacters(languageBook string) ([]Character, error) {
	return l.LoadCharactersContext(context.Background(), languageBook)
}

func (l Loader) LoadCharactersContext(ctx context.Context, languageBook string) ([]Character, error) {
	catalog, err := l.loadLocalizedCatalog(ctx, languageBook)
	if err != nil {
		return nil, err
	}
	return cloneCharacters(catalog.characters), nil
}

func (l Loader) LoadCharacter(languageBook string, id int64) (Character, bool, error) {
	return l.LoadCharacterContext(context.Background(), languageBook, id)
}

func (l Loader) LoadCharacterContext(ctx context.Context, languageBook string, id int64) (Character, bool, error) {
	catalog, err := l.loadLocalizedCatalog(ctx, languageBook)
	if err != nil {
		return Character{}, false, err
	}
	index, ok := catalog.indexByID[id]
	if !ok {
		return Character{}, false, nil
	}
	return cloneCharacter(catalog.characters[index]), true, nil
}

func (l Loader) loadLocalizedCatalog(ctx context.Context, languageBook string) (*localizedCatalog, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := ValidateBookName(languageBook); err != nil {
		return nil, err
	}
	baseFP, err := l.statBaseBooks()
	if err != nil {
		return nil, err
	}
	textFP, err := l.statBook(languageBook)
	if err != nil {
		return nil, err
	}

	if l.cache == nil {
		base, err := l.loadBaseDataset(ctx, baseFP)
		if err != nil {
			return nil, err
		}
		return l.buildLocalizedCatalog(base, languageBook, textFP)
	}

	l.cache.mu.RLock()
	cached := l.cache.catalogs[languageBook]
	l.cache.mu.RUnlock()
	if cached != nil && cached.baseFingerprint == baseFP && cached.textFingerprint == textFP {
		return cached, nil
	}

	resCh := l.cache.catalogGroup.DoChan(languageBook, func() (any, error) {
		freshBaseFP, err := l.statBaseBooks()
		if err != nil {
			return nil, err
		}
		freshTextFP, err := l.statBook(languageBook)
		if err != nil {
			return nil, err
		}

		l.cache.mu.RLock()
		curCatalog := l.cache.catalogs[languageBook]
		curBase := l.cache.base
		l.cache.mu.RUnlock()
		if curCatalog != nil && curCatalog.baseFingerprint == freshBaseFP && curCatalog.textFingerprint == freshTextFP {
			return curCatalog, nil
		}

		if curBase == nil || curBase.fingerprint != freshBaseFP {
			baseResult, err, _ := l.cache.baseGroup.Do("base", func() (any, error) {
				latestBaseFP, err := l.statBaseBooks()
				if err != nil {
					return nil, err
				}
				l.cache.mu.RLock()
				latestBase := l.cache.base
				l.cache.mu.RUnlock()
				if latestBase != nil && latestBase.fingerprint == latestBaseFP {
					return latestBase, nil
				}
				loaded, err := l.loadBaseDataset(context.Background(), latestBaseFP)
				if err != nil {
					return nil, err
				}
				l.cache.mu.Lock()
				l.cache.base = loaded
				l.cache.mu.Unlock()
				return loaded, nil
			})
			if err != nil {
				return nil, err
			}
			curBase = baseResult.(*baseDataset)
		}

		catalog, err := l.buildLocalizedCatalog(curBase, languageBook, freshTextFP)
		if err != nil {
			return nil, err
		}

		l.cache.mu.Lock()
		l.cache.base = curBase
		l.cache.catalogs[languageBook] = catalog
		l.cache.mu.Unlock()
		return catalog, nil
	})

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case res := <-resCh:
		if res.Err != nil {
			return nil, res.Err
		}
		return res.Val.(*localizedCatalog), nil
	}
}

func (l Loader) statBaseBooks() (baseDatasetFingerprint, error) {
	var fp baseDatasetFingerprint
	for i, name := range baseBookNames {
		bookFP, err := l.statBook(name)
		if err != nil {
			return baseDatasetFingerprint{}, err
		}
		fp[i] = bookFP
	}
	return fp, nil
}

func (l Loader) statBook(bookName string) (fileFingerprint, error) {
	info, err := os.Stat(filepath.Join(l.Root, bookName))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fileFingerprint{}, nil
		}
		return fileFingerprint{}, fmt.Errorf("stat master book %q: %w", bookName, err)
	}
	return fileFingerprint{
		exists:          true,
		size:            info.Size(),
		modTimeUnixNano: info.ModTime().UnixNano(),
	}, nil
}

func (l Loader) loadBaseDataset(ctx context.Context, fp baseDatasetFingerprint) (*baseDataset, error) {
	var (
		characterRows         []map[string]any
		profiles              map[int64]map[string]any
		storyRows             []map[string]any
		voiceRows             []map[string]any
		activeSkills          map[int64]map[string]any
		passiveSkills         map[int64]map[string]any
		uniqueWeapons         map[int64]map[string]any
		exclusiveEffects      map[int64]map[string]any
		exclusiveDescriptions map[int64]map[string]any
	)

	g, groupCtx := errgroup.WithContext(ctx)
	load := func(fn func() error) {
		g.Go(func() error {
			if err := groupCtx.Err(); err != nil {
				return err
			}
			return fn()
		})
	}

	load(func() error {
		rows, err := l.LoadRows("CharacterMB")
		characterRows = rows
		return err
	})
	load(func() error {
		rows, err := l.loadOptionalRows("CharacterProfileMB")
		if err != nil {
			return err
		}
		profiles = indexRowsByID(rows)
		return nil
	})
	load(func() error {
		rows, err := l.loadOptionalRows("CharacterStoryMB")
		storyRows = rows
		return err
	})
	load(func() error {
		rows, err := l.loadOptionalRows("CharacterDetailVoiceMB")
		voiceRows = rows
		return err
	})
	load(func() error {
		rows, err := l.loadOptionalRows("ActiveSkillMB")
		if err != nil {
			return err
		}
		activeSkills = indexRowsByID(rows)
		return nil
	})
	load(func() error {
		rows, err := l.loadOptionalRows("PassiveSkillMB")
		if err != nil {
			return err
		}
		passiveSkills = indexRowsByID(rows)
		return nil
	})
	load(func() error {
		rows, err := l.loadOptionalRows("EquipmentMB")
		if err != nil {
			return err
		}
		uniqueWeapons = indexUniqueWeapons(rows)
		return nil
	})
	load(func() error {
		rows, err := l.loadOptionalRows("EquipmentExclusiveEffectMB")
		if err != nil {
			return err
		}
		exclusiveEffects = indexRowsByID(rows)
		return nil
	})
	load(func() error {
		rows, err := l.loadOptionalRows("EquipmentExclusiveSkillDescriptionMB")
		if err != nil {
			return err
		}
		exclusiveDescriptions = indexRowsByID(rows)
		return nil
	})

	if err := g.Wait(); err != nil {
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	return &baseDataset{
		fingerprint:           fp,
		characterRows:         characterRows,
		profiles:              profiles,
		storyRows:             storyRows,
		voiceRows:             voiceRows,
		activeSkills:          activeSkills,
		passiveSkills:         passiveSkills,
		uniqueWeapons:         uniqueWeapons,
		exclusiveEffects:      exclusiveEffects,
		exclusiveDescriptions: exclusiveDescriptions,
	}, nil
}

func (l Loader) buildLocalizedCatalog(base *baseDataset, languageBook string, textFP fileFingerprint) (*localizedCatalog, error) {
	texts, err := l.LoadTextResources(languageBook)
	if err != nil {
		return nil, fmt.Errorf("load text resources: %w", err)
	}
	stories := groupStories(base.storyRows, texts)
	voices := groupVoices(base.voiceRows, texts)

	characters := make([]Character, 0, len(base.characterRows))
	indexByID := make(map[int64]int, len(base.characterRows))
	for _, row := range base.characterRows {
		id := asInt64(row["Id"])
		nameKey := asString(row["NameKey"])
		name2Key := asString(row["Name2Key"])
		profileRow := base.profiles[id]
		assets := characterAssets(id, asInt64(row["ElementType"]), asInt64(row["JobFlags"]))
		characterStories := stories[id]
		if characterStories == nil {
			characterStories = []CharacterStory{}
		}
		characterVoices := voices[id]
		if characterVoices == nil {
			characterVoices = []CharacterVoice{}
		}
		if _, exists := indexByID[id]; !exists {
			indexByID[id] = len(characters)
		}
		characters = append(characters, Character{
			ID:              id,
			Memo:            asString(row["Memo"]),
			Name:            textOrKey(texts, nameKey),
			Title:           textOrKey(texts, name2Key),
			NameKey:         nameKey,
			Name2Key:        name2Key,
			RarityFlags:     asInt64(row["RarityFlags"]),
			ElementType:     asInt64(row["ElementType"]),
			JobFlags:        asInt64(row["JobFlags"]),
			CharacterType:   asInt64(row["CharacterType"]),
			StartTimeFixJST: asString(row["StartTimeFixJST"]),
			EndTimeFixJST:   asString(row["EndTimeFixJST"]),
			Profile:         profileFromRow(profileRow, texts, id),
			Stats: CharacterStats{
				Speed:                         asInt64(nestedValue(row["InitialBattleParameter"], "Speed")),
				NormalSkillID:                 asInt64(row["NormalSkillId"]),
				ActiveSkillIDs:                asInt64Slice(row["ActiveSkillIds"]),
				PassiveSkillIDs:               asInt64Slice(row["PassiveSkillIds"]),
				BaseParameterCoefficient:      asInt64Map(row["BaseParameterCoefficient"]),
				BaseParameterGrossCoefficient: asInt64(row["BaseParameterGrossCoefficient"]),
				InitialBattleParameter:        asInt64Map(row["InitialBattleParameter"]),
			},
			Assets:       assets,
			Skills:       characterSkillsFromRow(row, base.activeSkills, base.passiveSkills, texts),
			UniqueWeapon: uniqueWeaponFromProfile(profileRow, base.uniqueWeapons, base.exclusiveEffects, base.exclusiveDescriptions, texts),
			Stories:      characterStories,
			Voices:       characterVoices,
			IconPath:     assets.IconSmall,
		})
	}
	return &localizedCatalog{
		baseFingerprint: base.fingerprint,
		textFingerprint: textFP,
		characters:      characters,
		indexByID:       indexByID,
	}, nil
}

func (l Loader) loadOptionalRows(bookName string) ([]map[string]any, error) {
	rows, err := l.LoadRows(bookName)
	if err == nil {
		return rows, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	return nil, err
}

func cloneCharacters(src []Character) []Character {
	out := make([]Character, len(src))
	for i, c := range src {
		out[i] = cloneCharacter(c)
	}
	return out
}

func cloneCharacter(c Character) Character {
	c.Profile.MovieJPURLs = cloneStringMap(c.Profile.MovieJPURLs)
	c.Profile.MovieUSURLs = cloneStringMap(c.Profile.MovieUSURLs)
	c.Stats.ActiveSkillIDs = cloneInt64Slice(c.Stats.ActiveSkillIDs)
	c.Stats.PassiveSkillIDs = cloneInt64Slice(c.Stats.PassiveSkillIDs)
	c.Stats.BaseParameterCoefficient = cloneInt64Map(c.Stats.BaseParameterCoefficient)
	c.Stats.InitialBattleParameter = cloneInt64Map(c.Stats.InitialBattleParameter)
	c.Skills.Normal = cloneSkill(c.Skills.Normal)
	c.Skills.Active = cloneSkills(c.Skills.Active)
	c.Skills.Passive = cloneSkills(c.Skills.Passive)
	c.UniqueWeapon.BattleParameter = cloneInt64Map(c.UniqueWeapon.BattleParameter)
	if c.UniqueWeapon.BonusParameters != nil {
		c.UniqueWeapon.BonusParameters = slices.Clone(c.UniqueWeapon.BonusParameters)
	}
	stories := make([]CharacterStory, len(c.Stories))
	for i, s := range c.Stories {
		if s.RewardItems != nil {
			s.RewardItems = slices.Clone(s.RewardItems)
		}
		s.AudioJP = cloneStringSlice(s.AudioJP)
		s.AudioUS = cloneStringSlice(s.AudioUS)
		stories[i] = s
	}
	c.Stories = stories
	c.Voices = slices.Clone(c.Voices)
	return c
}

func cloneSkills(src []CharacterSkill) []CharacterSkill {
	out := make([]CharacterSkill, len(src))
	for i, s := range src {
		out[i] = cloneSkill(s)
	}
	return out
}

func cloneSkill(s CharacterSkill) CharacterSkill {
	levels := make([]SkillLevel, len(s.Levels))
	for i, lv := range s.Levels {
		lv.SubSkillIDs = cloneInt64Slice(lv.SubSkillIDs)
		levels[i] = lv
	}
	s.Levels = levels
	return s
}

func cloneInt64Slice(src []int64) []int64 {
	if src == nil {
		return []int64{}
	}
	return slices.Clone(src)
}

func cloneStringSlice(src []string) []string {
	if src == nil {
		return []string{}
	}
	return slices.Clone(src)
}

func cloneInt64Map(src map[string]int64) map[string]int64 {
	if src == nil {
		return map[string]int64{}
	}
	return maps.Clone(src)
}

func cloneStringMap(src map[string]string) map[string]string {
	if src == nil {
		return map[string]string{}
	}
	return maps.Clone(src)
}

func profileFromRow(row map[string]any, texts map[string]string, id int64) Profile {
	if row == nil {
		return Profile{MovieJPURLs: map[string]string{}, MovieUSURLs: map[string]string{}}
	}
	voiceUSKey := asString(row["CharacterVoiceUSKey"])
	if voiceUSKey == "" && id != 0 {
		candidate := fmt.Sprintf("[CharacterProfileVoiceUS%d]", id)
		if _, ok := texts[candidate]; ok {
			voiceUSKey = candidate
		}
	}
	vocalUSKey := asString(row["VocalUSKey"])
	if vocalUSKey == "" && id != 0 {
		candidate := fmt.Sprintf("[CharacterProfileVocalUS%d]", id)
		if _, ok := texts[candidate]; ok {
			vocalUSKey = candidate
		}
	}
	return Profile{
		Birthday:     asInt64(row["Birthday"]),
		Height:       asInt64(row["Height"]),
		Weight:       asInt64(row["Weight"]),
		BloodType:    asInt64(row["BloodType"]),
		Description:  textOrKey(texts, asString(row["DescriptionKey"])),
		VoiceJP:      textOrKey(texts, asString(row["CharacterVoiceJPKey"])),
		VoiceUS:      textOrKey(texts, voiceUSKey),
		VocalJP:      textOrKey(texts, asString(row["VocalJPKey"])),
		VocalUS:      textOrKey(texts, vocalUSKey),
		LamentJP:     textOrKey(texts, asString(row["LamentJPKey"])),
		LamentUS:     textOrKey(texts, asString(row["LamentUSKey"])),
		LyricsJP:     textOrKey(texts, asString(row["LyricsJPKey"])),
		LyricsUS:     textOrKey(texts, asString(row["LyricsUSKey"])),
		GachaMessage: textOrKey(texts, asString(row["GachaResultMessage2Key"])),
		MovieJPURLs:  asStringMap(row["MovieJpUrl"]),
		MovieUSURLs:  asStringMap(row["MovieUsUrl"]),
	}
}

func groupStories(rows []map[string]any, texts map[string]string) map[int64][]CharacterStory {
	grouped := make(map[int64][]CharacterStory)
	for _, row := range rows {
		characterID := asInt64(row["CharacterId"])
		titleKey := asString(row["TitleKey"])
		textKey := asString(row["TextKey"])
		grouped[characterID] = append(grouped[characterID], CharacterStory{
			ID:          asInt64(row["Id"]),
			EpisodeID:   asInt64(row["EpisodeId"]),
			Level:       asInt64(row["Level"]),
			RarityFlags: asInt64(row["RarityFlags"]),
			Title:       textOrKey(texts, titleKey),
			Text:        textOrKey(texts, textKey),
			TitleKey:    titleKey,
			TextKey:     textKey,
			RewardItems: asRewardItems(row["RewardItemList"]),
			AudioJP:     []string{},
			AudioUS:     []string{},
		})
	}
	for id := range grouped {
		slices.SortStableFunc(grouped[id], func(a, b CharacterStory) int {
			return cmp.Or(
				cmp.Compare(a.EpisodeID, b.EpisodeID),
				cmp.Compare(a.ID, b.ID),
			)
		})
	}
	return grouped
}

func groupVoices(rows []map[string]any, texts map[string]string) map[int64][]CharacterVoice {
	grouped := make(map[int64][]CharacterVoice)
	for _, row := range rows {
		characterID := asInt64(row["CharacterId"])
		subtitleKey := asString(row["SubtitleKey"])
		buttonTextKey := asString(row["UnlockedVoiceButtonTextKey"])
		grouped[characterID] = append(grouped[characterID], CharacterVoice{
			ID:              asInt64(row["Id"]),
			Category:        asInt64(row["CharacterVoiceCategory"]),
			UnlockCondition: asInt64(row["UnlockCondition"]),
			UnlockQuestID:   asInt64(row["UnlockQuestId"]),
			SortOrder:       asInt64(row["SortOrder"]),
			Subtitle:        textOrKey(texts, subtitleKey),
			SubtitleKey:     subtitleKey,
			ButtonText:      textOrKey(texts, buttonTextKey),
			ButtonTextKey:   buttonTextKey,
		})
	}
	for id := range grouped {
		slices.SortStableFunc(grouped[id], func(a, b CharacterVoice) int {
			return cmp.Or(
				cmp.Compare(a.SortOrder, b.SortOrder),
				cmp.Compare(a.ID, b.ID),
			)
		})
	}
	return grouped
}

func characterAssets(id int64, elementType int64, jobFlags int64) CharacterAssets {
	baseIcon := fmt.Sprintf("/assets/Assets/AddressableConvertAssets/CharacterIcon/CHR_%06d/CHR_%06d_00", id, id)
	return CharacterAssets{
		IconSmall:     baseIcon + "_s.png",
		IconMedium:    baseIcon + "_m.png",
		IconLarge:     baseIcon + "_l.png",
		IconWide:      baseIcon + "_w.png",
		IconEmotion:   baseIcon + "_em_001.png",
		ElementIcon:   fmt.Sprintf("/assets/Assets/AddressableLocalAssets/Atlas/icon_element_%d.png", elementType),
		JobIcon:       fmt.Sprintf("/assets/Assets/AddressableLocalAssets/Atlas/icon_job_%s.png", jobSlug(jobFlags)),
		ModelTexture:  fmt.Sprintf("/assets/Assets/AddressableConvertAssets/Characters/CHR_%06d/model/CHR_%06d.2048/texture_00.png", id, id),
		EffectTexture: fmt.Sprintf("/assets/Assets/AddressableConvertAssets/Characters/CHR_%06d/effect/texture/CFX_%06d.png", id, id),
		SongJP:        fmt.Sprintf("/assets/Assets/AddressableConvertAssets/CharacterSong/CHR_%06d/CHR_%06d_SONG_JP.wav", id, id),
		SongUS:        fmt.Sprintf("/assets/Assets/AddressableConvertAssets/CharacterSong/CHR_%06d/CHR_%06d_SONG_US.wav", id, id),
		SongShortJP:   fmt.Sprintf("/assets/Assets/AddressableConvertAssets/CharacterSong/CHR_%06d/CHR_%06d_SONG_SHORT_JP.wav", id, id),
		SongShortUS:   fmt.Sprintf("/assets/Assets/AddressableConvertAssets/CharacterSong/CHR_%06d/CHR_%06d_SONG_SHORT_US.wav", id, id),
	}
}

func jobSlug(jobFlags int64) string {
	switch jobFlags {
	case 1:
		return "warrior"
	case 2:
		return "sniper"
	case 4:
		return "sorcerer"
	default:
		return fmt.Sprintf("%d", jobFlags)
	}
}

func textOrKey(texts map[string]string, key string) string {
	if key == "" {
		return ""
	}
	if value, ok := texts[key]; ok {
		return value
	}
	return key
}

func nestedValue(value any, key string) any {
	switch v := value.(type) {
	case map[string]any:
		return v[key]
	case map[any]any:
		return v[key]
	default:
		return nil
	}
}

func asString(value any) string {
	switch v := value.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	default:
		return ""
	}
}

func asInt64(value any) int64 {
	switch v := value.(type) {
	case int:
		return int64(v)
	case int8:
		return int64(v)
	case int16:
		return int64(v)
	case int32:
		return int64(v)
	case int64:
		return v
	case uint:
		return int64(v)
	case uint8:
		return int64(v)
	case uint16:
		return int64(v)
	case uint32:
		return int64(v)
	case uint64:
		if v > uint64(^uint(0)>>1) {
			return 0
		}
		return int64(v)
	default:
		return 0
	}
}

func asInt64Slice(value any) []int64 {
	switch v := value.(type) {
	case []int64:
		return v
	case []int:
		items := make([]int64, 0, len(v))
		for _, item := range v {
			items = append(items, int64(item))
		}
		return items
	case []any:
		items := make([]int64, 0, len(v))
		for _, item := range v {
			items = append(items, asInt64(item))
		}
		return items
	default:
		return []int64{}
	}
}

func asInt64Map(value any) map[string]int64 {
	out := map[string]int64{}
	switch v := value.(type) {
	case map[string]any:
		for key, item := range v {
			out[key] = asInt64(item)
		}
	case map[any]any:
		for key, item := range v {
			if stringKey, ok := key.(string); ok {
				out[stringKey] = asInt64(item)
			}
		}
	}
	return out
}

func asStringMap(value any) map[string]string {
	out := map[string]string{}
	switch v := value.(type) {
	case map[string]any:
		for key, item := range v {
			out[key] = asString(item)
		}
	case map[any]any:
		for key, item := range v {
			if stringKey, ok := key.(string); ok {
				out[stringKey] = asString(item)
			}
		}
	}
	return out
}

func asRewardItems(value any) []RewardItem {
	var rawItems []any
	switch v := value.(type) {
	case []any:
		rawItems = v
	case []map[string]any:
		rawItems = make([]any, 0, len(v))
		for _, item := range v {
			rawItems = append(rawItems, item)
		}
	default:
		return []RewardItem{}
	}
	items := make([]RewardItem, 0, len(rawItems))
	for _, raw := range rawItems {
		switch item := raw.(type) {
		case map[string]any:
			items = append(items, RewardItem{
				ItemID:    asInt64(item["ItemId"]),
				ItemType:  asInt64(item["ItemType"]),
				ItemCount: asInt64(item["ItemCount"]),
			})
		case map[any]any:
			items = append(items, RewardItem{
				ItemID:    asInt64(item["ItemId"]),
				ItemType:  asInt64(item["ItemType"]),
				ItemCount: asInt64(item["ItemCount"]),
			})
		}
	}
	return items
}
