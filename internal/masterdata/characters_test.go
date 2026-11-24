package masterdata

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/vmihailenco/msgpack/v5"
)

func TestLoadCharactersIncludesRichCharacterData(t *testing.T) {
	root := t.TempDir()
	writeMasterBook(t, root, "CharacterMB", []map[string]any{
		{
			"Id":                            int64(101),
			"Memo":                          "test character",
			"NameKey":                       "character.name.101",
			"Name2Key":                      "character.title.101",
			"RarityFlags":                   int64(4),
			"ElementType":                   int64(2),
			"JobFlags":                      int64(8),
			"CharacterType":                 int64(1),
			"StartTimeFixJST":               "2024-01-01 00:00:00",
			"EndTimeFixJST":                 "2100-12-31 23:59:59",
			"NormalSkillId":                 int64(10101),
			"ActiveSkillIds":                []any{int64(10102), int64(10103)},
			"PassiveSkillIds":               []any{int64(10104)},
			"BaseParameterCoefficient":      map[string]any{"Muscle": int64(90), "Energy": int64(80)},
			"BaseParameterGrossCoefficient": int64(341),
			"InitialBattleParameter":        map[string]any{"Speed": int64(2733), "Defense": int64(12)},
		},
	})
	writeMasterBook(t, root, "CharacterProfileMB", []map[string]any{
		{
			"Id":                     int64(101),
			"Birthday":               int64(307),
			"Height":                 int64(162),
			"Weight":                 int64(45),
			"BloodType":              int64(1),
			"DescriptionKey":         "character.description.101",
			"CharacterVoiceJPKey":    "character.voice.jp.101",
			"CharacterVoiceUSKey":    nil,
			"VocalJPKey":             "character.vocal.jp.101",
			"VocalUSKey":             nil,
			"LamentJPKey":            "character.lament.jp.101",
			"LamentUSKey":            "character.lament.us.101",
			"LyricsJPKey":            "character.lyrics.jp.101",
			"LyricsUSKey":            "character.lyrics.us.101",
			"GachaResultMessage2Key": "character.gacha.101",
			"EquipmentCompositeId":   int64(33),
			"MovieJpUrl":             map[string]any{"jaJP": "https://youtu.be/jp"},
			"MovieUsUrl":             map[string]any{"enUS": "https://youtu.be/us"},
		},
	})
	writeMasterBook(t, root, "CharacterStoryMB", []map[string]any{
		{
			"Id":             int64(501),
			"CharacterId":    int64(101),
			"EpisodeId":      int64(2),
			"Level":          int64(10),
			"RarityFlags":    int64(4),
			"TitleKey":       "story.title.2",
			"TextKey":        "story.text.2",
			"RewardItemList": []any{map[string]any{"ItemId": int64(1), "ItemType": int64(2), "ItemCount": int64(20)}},
		},
		{
			"Id":          int64(500),
			"CharacterId": int64(101),
			"EpisodeId":   int64(1),
			"TitleKey":    "story.title.1",
			"TextKey":     "story.text.1",
		},
	})
	writeMasterBook(t, root, "ActiveSkillMB", []map[string]any{
		{
			"Id":                    int64(10101),
			"NameKey":               "skill.normal.name",
			"ActiveSkillConditions": "True",
			"SkillInitCoolTime":     int64(0),
			"SkillMaxCoolTime":      int64(0),
			"ActiveSkillInfos": []any{map[string]any{
				"OrderNumber":          int64(1),
				"CharacterLevel":       int64(1),
				"EquipmentRarityFlags": int64(0),
				"DescriptionKey":       "skill.normal.description",
				"SubSetSkillIds":       []any{int64(999)},
			}},
		},
		{
			"Id":                    int64(10102),
			"NameKey":               "skill.active.name",
			"SkillInitCoolTime":     int64(1),
			"SkillMaxCoolTime":      int64(4),
			"ActiveSkillInfos":      []any{map[string]any{"OrderNumber": int64(1), "CharacterLevel": int64(11), "DescriptionKey": "skill.active.description"}},
			"ActiveSkillConditions": "HP > 0",
		},
	})
	writeMasterBook(t, root, "PassiveSkillMB", []map[string]any{
		{
			"Id":                int64(10104),
			"NameKey":           "skill.passive.name",
			"PassiveSkillInfos": []any{map[string]any{"OrderNumber": int64(1), "CharacterLevel": int64(41), "DescriptionKey": "skill.passive.description"}},
		},
	})
	writeMasterBook(t, root, "EquipmentMB", []map[string]any{
		{
			"Id":                                   int64(9001),
			"NameKey":                              "equipment.unique.name",
			"IconId":                               int64(777),
			"EquippedJobFlags":                     int64(8),
			"RarityFlags":                          int64(128),
			"EquipmentLv":                          int64(180),
			"SlotType":                             int64(1),
			"CompositeId":                          int64(33),
			"ExclusiveEffectId":                    int64(44),
			"EquipmentExclusiveSkillDescriptionId": int64(55),
			"BattleParameterChangeInfo":            map[string]any{"BattleParameterType": int64(2), "ChangeParameterType": int64(1), "Value": int64(500)},
		},
	})
	writeMasterBook(t, root, "EquipmentExclusiveEffectMB", []map[string]any{
		{
			"Id":                          int64(44),
			"CharacterId":                 int64(101),
			"BaseParameterChangeInfoList": []any{map[string]any{"BaseParameterType": int64(1), "ChangeParameterType": int64(2), "Value": int64(1200)}},
		},
	})
	writeMasterBook(t, root, "EquipmentExclusiveSkillDescriptionMB", []map[string]any{
		{"Id": int64(55), "Description1Key": "equipment.unique.ssr", "Description2Key": "equipment.unique.ur", "Description3Key": "equipment.unique.lr"},
	})
	writeMasterBook(t, root, "CharacterDetailVoiceMB", []map[string]any{
		{
			"Id":                         int64(301),
			"CharacterId":                int64(101),
			"CharacterVoiceCategory":     int64(3),
			"UnlockCondition":            int64(0),
			"UnlockQuestId":              int64(0),
			"SortOrder":                  int64(8),
			"SubtitleKey":                "voice.subtitle.2",
			"UnlockedVoiceButtonTextKey": "voice.button.2",
		},
		{
			"Id":                         int64(300),
			"CharacterId":                int64(101),
			"CharacterVoiceCategory":     int64(2),
			"SortOrder":                  int64(1),
			"SubtitleKey":                "voice.subtitle.1",
			"UnlockedVoiceButtonTextKey": "voice.button.1",
		},
	})
	writeMasterBook(t, root, "TextResourceEnUsMB", []map[string]any{
		{"StringKey": "character.name.101", "Text": "Iris"},
		{"StringKey": "character.title.101", "Text": "The Bloom"},
		{"StringKey": "character.description.101", "Text": "A gentle singer."},
		{"StringKey": "character.voice.jp.101", "Text": "Kana Voice"},
		{"StringKey": "[CharacterProfileVoiceUS101]", "Text": "Ava Voice"},
		{"StringKey": "character.vocal.jp.101", "Text": "Singer JP"},
		{"StringKey": "[CharacterProfileVocalUS101]", "Text": "Singer US"},
		{"StringKey": "character.lament.jp.101", "Text": "JP Lament"},
		{"StringKey": "character.lament.us.101", "Text": "US Lament"},
		{"StringKey": "character.lyrics.jp.101", "Text": "JP lyrics"},
		{"StringKey": "character.lyrics.us.101", "Text": "US lyrics"},
		{"StringKey": "character.gacha.101", "Text": "I will join you."},
		{"StringKey": "story.title.1", "Text": "First memory"},
		{"StringKey": "story.text.1", "Text": "The first story."},
		{"StringKey": "story.title.2", "Text": "Second memory"},
		{"StringKey": "story.text.2", "Text": "The second story."},
		{"StringKey": "voice.subtitle.1", "Text": "Good morning."},
		{"StringKey": "voice.button.1", "Text": "Greeting"},
		{"StringKey": "voice.subtitle.2", "Text": "Let's fight."},
		{"StringKey": "voice.button.2", "Text": "Battle"},
		{"StringKey": "skill.normal.name", "Text": "Normal Attack"},
		{"StringKey": "skill.normal.description", "Text": "Deals magic damage."},
		{"StringKey": "skill.active.name", "Text": "Bloom Strike"},
		{"StringKey": "skill.active.description", "Text": "A bright attack."},
		{"StringKey": "skill.passive.name", "Text": "Quiet Resolve"},
		{"StringKey": "skill.passive.description", "Text": "Raises defense."},
		{"StringKey": "equipment.unique.name", "Text": "Iris Tome"},
		{"StringKey": "equipment.unique.ssr", "Text": "SSR weapon bonus."},
		{"StringKey": "equipment.unique.ur", "Text": "UR weapon bonus."},
		{"StringKey": "equipment.unique.lr", "Text": "LR weapon bonus."},
	})

	characters, err := New(root).LoadCharacters("TextResourceEnUsMB")
	if err != nil {
		t.Fatal(err)
	}
	if len(characters) != 1 {
		t.Fatalf("len(characters)=%d want 1", len(characters))
	}
	character := characters[0]
	if character.Name != "Iris" || character.Title != "The Bloom" {
		t.Fatalf("unexpected display names: %#v", character)
	}
	if character.Profile.Description != "A gentle singer." || character.Profile.VoiceUS != "Ava Voice" || character.Profile.LyricsUS != "US lyrics" {
		t.Fatalf("unexpected rich profile: %#v", character.Profile)
	}
	if character.Profile.MovieJPURLs["jaJP"] != "https://youtu.be/jp" || character.Profile.MovieUSURLs["enUS"] != "https://youtu.be/us" {
		t.Fatalf("unexpected movie urls: %#v %#v", character.Profile.MovieJPURLs, character.Profile.MovieUSURLs)
	}
	if character.Stats.Speed != 2733 || character.Stats.NormalSkillID != 10101 || character.Stats.BaseParameterCoefficient["Muscle"] != 90 {
		t.Fatalf("unexpected stats: %#v", character.Stats)
	}
	if len(character.Stats.ActiveSkillIDs) != 2 || character.Stats.ActiveSkillIDs[1] != 10103 {
		t.Fatalf("unexpected active skills: %#v", character.Stats.ActiveSkillIDs)
	}
	if character.Assets.IconLarge != "/assets/Assets/AddressableConvertAssets/CharacterIcon/CHR_000101/CHR_000101_00_l.png" {
		t.Fatalf("unexpected large icon: %q", character.Assets.IconLarge)
	}
	if character.IconPath != character.Assets.IconSmall {
		t.Fatalf("iconPath=%q assets.iconSmall=%q", character.IconPath, character.Assets.IconSmall)
	}
	if len(character.Stories) != 2 || character.Stories[0].EpisodeID != 1 || character.Stories[1].RewardItems[0].ItemCount != 20 {
		t.Fatalf("stories not localized/sorted: %#v", character.Stories)
	}
	if len(character.Voices) != 2 || character.Voices[0].ButtonText != "Greeting" || character.Voices[1].Subtitle != "Let's fight." {
		t.Fatalf("voices not localized/sorted: %#v", character.Voices)
	}
	if character.Skills.Normal.Name != "Normal Attack" || len(character.Skills.Active) != 1 || character.Skills.Active[0].MaxCooltime != 4 || character.Skills.Active[0].Levels[0].Description != "A bright attack." {
		t.Fatalf("skills not localized/projected: %#v", character.Skills)
	}
	if len(character.Skills.Passive) != 1 || character.Skills.Passive[0].Name != "Quiet Resolve" {
		t.Fatalf("passive skills not localized/projected: %#v", character.Skills.Passive)
	}
	if character.UniqueWeapon.Name != "Iris Tome" || character.UniqueWeapon.Icon != "/assets/Assets/AddressableConvertAssets/Icon/Equipment/EQP_000777.png" || character.UniqueWeapon.Descriptions.LR != "LR weapon bonus." {
		t.Fatalf("unique weapon not localized/projected: %#v", character.UniqueWeapon)
	}
	if len(character.UniqueWeapon.BonusParameters) != 1 || character.UniqueWeapon.BonusParameters[0].Value != 1200 {
		t.Fatalf("unique weapon bonus missing: %#v", character.UniqueWeapon.BonusParameters)
	}
}

func TestLoadCharactersToleratesMissingOptionalCharacterBooks(t *testing.T) {
	root := t.TempDir()
	writeMasterBook(t, root, "CharacterMB", []map[string]any{{"Id": int64(7), "NameKey": "name.7"}})
	writeMasterBook(t, root, "TextResourceEnUsMB", []map[string]any{{"StringKey": "name.7", "Text": "Seven"}})

	characters, err := New(root).LoadCharacters("TextResourceEnUsMB")
	if err != nil {
		t.Fatal(err)
	}
	if len(characters) != 1 {
		t.Fatalf("len(characters)=%d want 1", len(characters))
	}
	character := characters[0]
	if character.Name != "Seven" || len(character.Stories) != 0 || len(character.Voices) != 0 {
		t.Fatalf("unexpected character with missing optional books: %#v", character)
	}
	if character.Stories == nil || character.Voices == nil || character.Stats.ActiveSkillIDs == nil || character.Stats.PassiveSkillIDs == nil || character.Skills.Active == nil || character.Skills.Passive == nil || character.UniqueWeapon.BonusParameters == nil {
		t.Fatalf("array fields must be non-nil for JSON clients: %#v", character)
	}
	if character.Profile.VoiceUS != "" || character.Profile.VocalUS != "" {
		t.Fatalf("missing profile should not synthesize unresolved text keys: %#v", character.Profile)
	}
	if character.Profile.MovieJPURLs == nil || character.Profile.MovieUSURLs == nil {
		t.Fatalf("profile url maps must be non-nil for JSON clients: %#v", character.Profile)
	}
}

func TestLoadCharacterAndCacheInvalidation(t *testing.T) {
	root := t.TempDir()
	writeMasterBook(t, root, "CharacterMB", []map[string]any{{"Id": int64(7), "NameKey": "name.7"}})
	writeMasterBook(t, root, "TextResourceEnUsMB", []map[string]any{{"StringKey": "name.7", "Text": "Seven"}})

	loader := New(root)

	got, found, err := loader.LoadCharacter("TextResourceEnUsMB", 7)
	if err != nil {
		t.Fatal(err)
	}
	if !found || got.Name != "Seven" {
		t.Fatalf("LoadCharacter(7) = (%#v, %v), want Name=Seven", got, found)
	}
	if _, found, err := loader.LoadCharacter("TextResourceEnUsMB", 999); err != nil || found {
		t.Fatalf("LoadCharacter(999) = (found=%v, err=%v), want (false, nil)", found, err)
	}

	writeMasterBook(t, root, "TextResourceEnUsMB", []map[string]any{{"StringKey": "name.7", "Text": "Seven Updated"}})
	future := time.Now().Add(2 * time.Second)
	if err := os.Chtimes(filepath.Join(root, "TextResourceEnUsMB"), future, future); err != nil {
		t.Fatal(err)
	}
	got, found, err = loader.LoadCharacter("TextResourceEnUsMB", 7)
	if err != nil {
		t.Fatal(err)
	}
	if !found || got.Name != "Seven Updated" {
		t.Fatalf("after text update LoadCharacter(7) = (%#v, %v), want Name=Seven Updated", got, found)
	}

	writeMasterBook(t, root, "CharacterProfileMB", []map[string]any{
		{"Id": int64(7), "DescriptionKey": "desc.7"},
	})
	writeMasterBook(t, root, "TextResourceEnUsMB", []map[string]any{
		{"StringKey": "name.7", "Text": "Seven Updated"},
		{"StringKey": "desc.7", "Text": "Added profile"},
	})
	future2 := future.Add(2 * time.Second)
	_ = os.Chtimes(filepath.Join(root, "CharacterProfileMB"), future2, future2)
	_ = os.Chtimes(filepath.Join(root, "TextResourceEnUsMB"), future2, future2)

	got, found, err = loader.LoadCharacter("TextResourceEnUsMB", 7)
	if err != nil {
		t.Fatal(err)
	}
	if !found || got.Profile.Description != "Added profile" {
		t.Fatalf("after base update LoadCharacter(7) profile = %#v, want Added profile", got.Profile)
	}

	canceledCtx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := loader.LoadCharactersContext(canceledCtx, "TextResourceEnUsMB"); !errors.Is(err, context.Canceled) {
		t.Fatalf("LoadCharactersContext(canceled) err = %v, want context.Canceled", err)
	}
}

func writeMasterBook(t *testing.T, root, name string, rows []map[string]any) {
	t.Helper()
	data, err := msgpack.Marshal(rows)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, name), data, 0o644); err != nil {
		t.Fatal(err)
	}
}
