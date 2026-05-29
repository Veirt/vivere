package masterdata

import (
	"errors"
	"fmt"
	"os"
	"sort"
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
	texts, err := l.LoadTextResources(languageBook)
	if err != nil {
		return nil, fmt.Errorf("load text resources: %w", err)
	}
	characterRows, err := l.LoadRows("CharacterMB")
	if err != nil {
		return nil, err
	}
	profileRows, err := l.loadOptionalRows("CharacterProfileMB")
	if err != nil {
		return nil, err
	}
	storyRows, err := l.loadOptionalRows("CharacterStoryMB")
	if err != nil {
		return nil, err
	}
	voiceRows, err := l.loadOptionalRows("CharacterDetailVoiceMB")
	if err != nil {
		return nil, err
	}
	activeSkillRows, err := l.loadOptionalRows("ActiveSkillMB")
	if err != nil {
		return nil, err
	}
	passiveSkillRows, err := l.loadOptionalRows("PassiveSkillMB")
	if err != nil {
		return nil, err
	}
	equipmentRows, err := l.loadOptionalRows("EquipmentMB")
	if err != nil {
		return nil, err
	}
	exclusiveEffectRows, err := l.loadOptionalRows("EquipmentExclusiveEffectMB")
	if err != nil {
		return nil, err
	}
	exclusiveDescriptionRows, err := l.loadOptionalRows("EquipmentExclusiveSkillDescriptionMB")
	if err != nil {
		return nil, err
	}

	profiles := make(map[int64]map[string]any, len(profileRows))
	for _, row := range profileRows {
		profiles[asInt64(row["Id"])] = row
	}
	stories := groupStories(storyRows, texts)
	voices := groupVoices(voiceRows, texts)
	activeSkills := indexRowsByID(activeSkillRows)
	passiveSkills := indexRowsByID(passiveSkillRows)
	uniqueWeapons := indexUniqueWeapons(equipmentRows)
	exclusiveEffects := indexRowsByID(exclusiveEffectRows)
	exclusiveDescriptions := indexRowsByID(exclusiveDescriptionRows)

	characters := make([]Character, 0, len(characterRows))
	for _, row := range characterRows {
		id := asInt64(row["Id"])
		nameKey := asString(row["NameKey"])
		name2Key := asString(row["Name2Key"])
		profileRow := profiles[id]
		assets := characterAssets(id, asInt64(row["ElementType"]), asInt64(row["JobFlags"]))
		characterStories := stories[id]
		if characterStories == nil {
			characterStories = []CharacterStory{}
		}
		characterVoices := voices[id]
		if characterVoices == nil {
			characterVoices = []CharacterVoice{}
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
			Skills:       characterSkillsFromRow(row, activeSkills, passiveSkills, texts),
			UniqueWeapon: uniqueWeaponFromProfile(profileRow, uniqueWeapons, exclusiveEffects, exclusiveDescriptions, texts),
			Stories:      characterStories,
			Voices:       characterVoices,
			IconPath:     assets.IconSmall,
		})
	}
	return characters, nil
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
		sort.SliceStable(grouped[id], func(i, j int) bool {
			if grouped[id][i].EpisodeID == grouped[id][j].EpisodeID {
				return grouped[id][i].ID < grouped[id][j].ID
			}
			return grouped[id][i].EpisodeID < grouped[id][j].EpisodeID
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
		sort.SliceStable(grouped[id], func(i, j int) bool {
			if grouped[id][i].SortOrder == grouped[id][j].SortOrder {
				return grouped[id][i].ID < grouped[id][j].ID
			}
			return grouped[id][i].SortOrder < grouped[id][j].SortOrder
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
