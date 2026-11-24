package masterdata

import "fmt"

type CharacterSkills struct {
	Normal  CharacterSkill   `json:"normal"`
	Active  []CharacterSkill `json:"active"`
	Passive []CharacterSkill `json:"passive"`
}

type CharacterSkill struct {
	ID           int64        `json:"id"`
	Type         string       `json:"type"`
	Name         string       `json:"name"`
	NameKey      string       `json:"nameKey"`
	Icon         string       `json:"icon"`
	Conditions   string       `json:"conditions"`
	InitCooltime int64        `json:"initCooltime"`
	MaxCooltime  int64        `json:"maxCooltime"`
	Levels       []SkillLevel `json:"levels"`
}

type SkillLevel struct {
	OrderNumber          int64   `json:"orderNumber"`
	CharacterLevel       int64   `json:"characterLevel"`
	EquipmentRarityFlags int64   `json:"equipmentRarityFlags"`
	Description          string  `json:"description"`
	DescriptionKey       string  `json:"descriptionKey"`
	SubSkillIDs          []int64 `json:"subSkillIds"`
}

type UniqueWeapon struct {
	EquipmentID        int64                    `json:"equipmentId"`
	Name               string                   `json:"name"`
	NameKey            string                   `json:"nameKey"`
	IconID             int64                    `json:"iconId"`
	Icon               string                   `json:"icon"`
	JobFlags           int64                    `json:"jobFlags"`
	RarityFlags        int64                    `json:"rarityFlags"`
	Level              int64                    `json:"level"`
	SlotType           int64                    `json:"slotType"`
	CompositeID        int64                    `json:"compositeId"`
	ExclusiveEffectID  int64                    `json:"exclusiveEffectId"`
	SkillDescriptionID int64                    `json:"skillDescriptionId"`
	BattleParameter    map[string]int64         `json:"battleParameter"`
	BonusParameters    []ParameterChange        `json:"bonusParameters"`
	Descriptions       UniqueWeaponDescriptions `json:"descriptions"`
}

type ParameterChange struct {
	Category   string `json:"category"`
	Type       int64  `json:"type"`
	ChangeType int64  `json:"changeType"`
	Value      int64  `json:"value"`
}

type UniqueWeaponDescriptions struct {
	SSR string `json:"ssr"`
	UR  string `json:"ur"`
	LR  string `json:"lr"`
}

func characterSkillsFromRow(row map[string]any, activeSkills map[int64]map[string]any, passiveSkills map[int64]map[string]any, texts map[string]string) CharacterSkills {
	activeIDs := asInt64Slice(row["ActiveSkillIds"])
	passiveIDs := asInt64Slice(row["PassiveSkillIds"])
	return CharacterSkills{
		Normal:  activeSkillFromRow(activeSkills[asInt64(row["NormalSkillId"])], texts),
		Active:  skillsFromIDs(activeIDs, activeSkills, "active", texts),
		Passive: skillsFromIDs(passiveIDs, passiveSkills, "passive", texts),
	}
}

func skillsFromIDs(ids []int64, rows map[int64]map[string]any, skillType string, texts map[string]string) []CharacterSkill {
	skills := make([]CharacterSkill, 0, len(ids))
	for _, id := range ids {
		row := rows[id]
		if row == nil {
			continue
		}
		if skillType == "passive" {
			skills = append(skills, passiveSkillFromRow(row, texts))
			continue
		}
		skills = append(skills, activeSkillFromRow(row, texts))
	}
	return skills
}

func activeSkillFromRow(row map[string]any, texts map[string]string) CharacterSkill {
	if row == nil {
		return CharacterSkill{Type: "active", Levels: []SkillLevel{}}
	}
	id := asInt64(row["Id"])
	nameKey := asString(row["NameKey"])
	return CharacterSkill{
		ID:           id,
		Type:         "active",
		Name:         textOrKey(texts, nameKey),
		NameKey:      nameKey,
		Icon:         skillIcon(id),
		Conditions:   asString(row["ActiveSkillConditions"]),
		InitCooltime: asInt64(row["SkillInitCoolTime"]),
		MaxCooltime:  asInt64(row["SkillMaxCoolTime"]),
		Levels:       skillLevelsFromRows(row["ActiveSkillInfos"], texts),
	}
}

func passiveSkillFromRow(row map[string]any, texts map[string]string) CharacterSkill {
	if row == nil {
		return CharacterSkill{Type: "passive", Levels: []SkillLevel{}}
	}
	id := asInt64(row["Id"])
	nameKey := asString(row["NameKey"])
	return CharacterSkill{
		ID:      id,
		Type:    "passive",
		Name:    textOrKey(texts, nameKey),
		NameKey: nameKey,
		Icon:    skillIcon(id),
		Levels:  skillLevelsFromRows(row["PassiveSkillInfos"], texts),
	}
}

func skillLevelsFromRows(value any, texts map[string]string) []SkillLevel {
	rawRows := asMapSlice(value)
	levels := make([]SkillLevel, 0, len(rawRows))
	for _, row := range rawRows {
		descriptionKey := asString(row["DescriptionKey"])
		levels = append(levels, SkillLevel{
			OrderNumber:          asInt64(row["OrderNumber"]),
			CharacterLevel:       asInt64(row["CharacterLevel"]),
			EquipmentRarityFlags: asInt64(row["EquipmentRarityFlags"]),
			Description:          textOrKey(texts, descriptionKey),
			DescriptionKey:       descriptionKey,
			SubSkillIDs:          asInt64Slice(row["SubSetSkillIds"]),
		})
	}
	return levels
}

func indexUniqueWeapons(rows []map[string]any) map[int64]map[string]any {
	uniqueWeapons := map[int64]map[string]any{}
	fallbacks := map[int64]map[string]any{}
	for _, row := range rows {
		compositeID := asInt64(row["CompositeId"])
		if compositeID == 0 || asInt64(row["ExclusiveEffectId"]) == 0 {
			continue
		}
		if _, ok := fallbacks[compositeID]; !ok {
			fallbacks[compositeID] = row
		}
		if asInt64(row["RarityFlags"]) == 128 && asInt64(row["EquipmentLv"]) == 180 {
			uniqueWeapons[compositeID] = row
		}
	}
	for compositeID, row := range fallbacks {
		if _, ok := uniqueWeapons[compositeID]; !ok {
			uniqueWeapons[compositeID] = row
		}
	}
	return uniqueWeapons
}

func uniqueWeaponFromProfile(profileRow map[string]any, uniqueWeapons map[int64]map[string]any, exclusiveEffects map[int64]map[string]any, exclusiveDescriptions map[int64]map[string]any, texts map[string]string) UniqueWeapon {
	if profileRow == nil {
		return emptyUniqueWeapon()
	}
	compositeID := asInt64(profileRow["EquipmentCompositeId"])
	row := uniqueWeapons[compositeID]
	if row == nil {
		return emptyUniqueWeapon()
	}
	nameKey := asString(row["NameKey"])
	descriptionID := asInt64(row["EquipmentExclusiveSkillDescriptionId"])
	effectID := asInt64(row["ExclusiveEffectId"])
	return UniqueWeapon{
		EquipmentID:        asInt64(row["Id"]),
		Name:               textOrKey(texts, nameKey),
		NameKey:            nameKey,
		IconID:             asInt64(row["IconId"]),
		Icon:               equipmentIcon(asInt64(row["IconId"])),
		JobFlags:           asInt64(row["EquippedJobFlags"]),
		RarityFlags:        asInt64(row["RarityFlags"]),
		Level:              asInt64(row["EquipmentLv"]),
		SlotType:           asInt64(row["SlotType"]),
		CompositeID:        compositeID,
		ExclusiveEffectID:  effectID,
		SkillDescriptionID: descriptionID,
		BattleParameter:    asInt64Map(row["BattleParameterChangeInfo"]),
		BonusParameters:    exclusiveBonusParameters(exclusiveEffects[effectID]),
		Descriptions:       uniqueWeaponDescriptions(exclusiveDescriptions[descriptionID], texts),
	}
}

func emptyUniqueWeapon() UniqueWeapon {
	return UniqueWeapon{
		BattleParameter: map[string]int64{},
		BonusParameters: []ParameterChange{},
	}
}

func uniqueWeaponDescriptions(row map[string]any, texts map[string]string) UniqueWeaponDescriptions {
	if row == nil {
		return UniqueWeaponDescriptions{}
	}
	return UniqueWeaponDescriptions{
		SSR: textOrKey(texts, asString(row["Description1Key"])),
		UR:  textOrKey(texts, asString(row["Description2Key"])),
		LR:  textOrKey(texts, asString(row["Description3Key"])),
	}
}

func exclusiveBonusParameters(row map[string]any) []ParameterChange {
	if row == nil {
		return []ParameterChange{}
	}
	changes := make([]ParameterChange, 0)
	changes = append(changes, parameterChangesFromRows(row["BaseParameterChangeInfoList"], "base")...)
	changes = append(changes, parameterChangesFromRows(row["BattleParameterChangeInfoList"], "battle")...)
	return changes
}

func parameterChangesFromRows(value any, category string) []ParameterChange {
	rows := asMapSlice(value)
	changes := make([]ParameterChange, 0, len(rows))
	for _, row := range rows {
		changeType := asInt64(row["ChangeParameterType"])
		parameterType := asInt64(row["BaseParameterType"])
		if category == "battle" {
			parameterType = asInt64(row["BattleParameterType"])
		}
		changes = append(changes, ParameterChange{
			Category:   category,
			Type:       parameterType,
			ChangeType: changeType,
			Value:      asInt64(row["Value"]),
		})
	}
	return changes
}

func indexRowsByID(rows []map[string]any) map[int64]map[string]any {
	indexed := make(map[int64]map[string]any, len(rows))
	for _, row := range rows {
		indexed[asInt64(row["Id"])] = row
	}
	return indexed
}

func asMapSlice(value any) []map[string]any {
	switch v := value.(type) {
	case []map[string]any:
		return v
	case []any:
		rows := make([]map[string]any, 0, len(v))
		for _, item := range v {
			switch row := item.(type) {
			case map[string]any:
				rows = append(rows, row)
			case map[any]any:
				converted := make(map[string]any, len(row))
				for key, value := range row {
					if stringKey, ok := key.(string); ok {
						converted[stringKey] = value
					}
				}
				rows = append(rows, converted)
			}
		}
		return rows
	default:
		return []map[string]any{}
	}
}

func skillIcon(id int64) string {
	if id == 0 {
		return ""
	}
	return fmt.Sprintf("/assets/Assets/AddressableConvertAssets/Icon/Skill/CSK_%09d.png", id)
}

func equipmentIcon(id int64) string {
	if id == 0 {
		return ""
	}
	return fmt.Sprintf("/assets/Assets/AddressableConvertAssets/Icon/Equipment/EQP_%06d.png", id)
}
