package addressables

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"unicode/utf16"
)

type Catalog struct {
	ProviderIDs []string `json:"m_ProviderIds"`
	InternalIDs []string `json:"m_InternalIds"`
	KeyData     string   `json:"m_KeyDataString"`
	BucketData  string   `json:"m_BucketDataString"`
	EntryData   string   `json:"m_EntryDataString"`
	ExtraData   string   `json:"m_ExtraDataString"`
	Resources   []ResourceLocation
}

type ResourceLocation struct {
	InternalID string
	ProviderID string
	Dependency any
	Data       any
	PrimaryKey string
}

type JSONClassObject struct {
	AssemblyName string
	ClassName    string
	JSONText     string
}

type bucket struct {
	offset  int32
	entries []int32
}

type entry struct {
	internalIDIndex    int32
	providerIndex      int32
	dependencyKeyIndex int32
	dependencyHash     int32
	dataIndex          int32
	primaryKeyIndex    int32
	resourceTypeIndex  int32
}

func Parse(data []byte) (Catalog, error) {
	var cat Catalog
	if err := json.Unmarshal(data, &cat); err != nil {
		return Catalog{}, err
	}
	resources, err := decodeResources(cat)
	if err != nil {
		return Catalog{}, err
	}
	cat.Resources = resources
	return cat, nil
}

func decodeResources(cat Catalog) ([]ResourceLocation, error) {
	buckets, err := decodeBuckets(cat.BucketData)
	if err != nil {
		return nil, err
	}
	keys, err := decodeKeys(cat.KeyData, buckets)
	if err != nil {
		return nil, err
	}
	entries, err := decodeEntries(cat.EntryData)
	if err != nil {
		return nil, err
	}
	extraData, err := base64.StdEncoding.DecodeString(cat.ExtraData)
	if err != nil {
		return nil, err
	}
	extraReader := bytes.NewReader(extraData)
	locations := make([]ResourceLocation, 0, len(entries))
	for _, ent := range entries {
		if ent.internalIDIndex < 0 || int(ent.internalIDIndex) >= len(cat.InternalIDs) {
			return nil, fmt.Errorf("invalid internal id index %d", ent.internalIDIndex)
		}
		if ent.providerIndex < 0 || int(ent.providerIndex) >= len(cat.ProviderIDs) {
			return nil, fmt.Errorf("invalid provider index %d", ent.providerIndex)
		}
		var dependency any
		if ent.dependencyKeyIndex >= 0 {
			if int(ent.dependencyKeyIndex) >= len(keys) {
				return nil, fmt.Errorf("invalid dependency index %d", ent.dependencyKeyIndex)
			}
			dependency = keys[ent.dependencyKeyIndex]
		}
		var objData any
		if ent.dataIndex >= 0 {
			if _, err := extraReader.Seek(int64(ent.dataIndex), io.SeekStart); err != nil {
				return nil, err
			}
			objData, err = decodeSerializedObject(extraReader)
			if err != nil {
				return nil, err
			}
		}
		if ent.primaryKeyIndex < 0 || int(ent.primaryKeyIndex) >= len(keys) {
			return nil, fmt.Errorf("invalid primary key index %d", ent.primaryKeyIndex)
		}
		locations = append(locations, ResourceLocation{
			InternalID: cat.InternalIDs[ent.internalIDIndex],
			ProviderID: cat.ProviderIDs[ent.providerIndex],
			Dependency: dependency,
			Data:       objData,
			PrimaryKey: fmt.Sprint(keys[ent.primaryKeyIndex]),
		})
	}
	return locations, nil
}

func decodeBuckets(encoded string) ([]bucket, error) {
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, err
	}
	r := bytes.NewReader(data)
	var count int32
	if err := binary.Read(r, binary.LittleEndian, &count); err != nil {
		return nil, err
	}
	if count < 0 {
		return nil, fmt.Errorf("invalid bucket count %d", count)
	}
	buckets := make([]bucket, 0, count)
	for range count {
		var b bucket
		if err := binary.Read(r, binary.LittleEndian, &b.offset); err != nil {
			return nil, err
		}
		var entryCount int32
		if err := binary.Read(r, binary.LittleEndian, &entryCount); err != nil {
			return nil, err
		}
		if entryCount < 0 {
			return nil, fmt.Errorf("invalid bucket entry count %d", entryCount)
		}
		b.entries = make([]int32, entryCount)
		for i := range b.entries {
			if err := binary.Read(r, binary.LittleEndian, &b.entries[i]); err != nil {
				return nil, err
			}
		}
		buckets = append(buckets, b)
	}
	return buckets, nil
}

func decodeKeys(encoded string, buckets []bucket) ([]any, error) {
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, err
	}
	r := bytes.NewReader(data)
	var keyCount int32
	if err := binary.Read(r, binary.LittleEndian, &keyCount); err != nil {
		return nil, err
	}
	if keyCount < 0 {
		return nil, fmt.Errorf("invalid key count %d", keyCount)
	}
	keys := make([]any, 0, keyCount)
	for i := int32(0); i < keyCount; i++ {
		if int(i) >= len(buckets) {
			return nil, fmt.Errorf("bucket missing for key %d", i)
		}
		if _, err := r.Seek(int64(buckets[i].offset), io.SeekStart); err != nil {
			return nil, err
		}
		v, err := decodeSerializedObject(r)
		if err != nil {
			return nil, err
		}
		keys = append(keys, v)
	}
	return keys, nil
}

func decodeEntries(encoded string) ([]entry, error) {
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, err
	}
	r := bytes.NewReader(data)
	var count int32
	if err := binary.Read(r, binary.LittleEndian, &count); err != nil {
		return nil, err
	}
	if count < 0 {
		return nil, fmt.Errorf("invalid entry count %d", count)
	}
	entries := make([]entry, 0, count)
	for range count {
		var e entry
		for _, field := range []*int32{&e.internalIDIndex, &e.providerIndex, &e.dependencyKeyIndex, &e.dependencyHash, &e.dataIndex, &e.primaryKeyIndex, &e.resourceTypeIndex} {
			if err := binary.Read(r, binary.LittleEndian, field); err != nil {
				return nil, err
			}
		}
		entries = append(entries, e)
	}
	return entries, nil
}

func decodeSerializedObject(r io.ReadSeeker) (any, error) {
	var objectType uint8
	if err := binary.Read(r, binary.LittleEndian, &objectType); err != nil {
		return nil, err
	}
	switch objectType {
	case 0:
		return readASCII32(r)
	case 1:
		return readUTF16LE32(r)
	case 2:
		var v uint16
		return v, binary.Read(r, binary.LittleEndian, &v)
	case 3:
		var v uint32
		return v, binary.Read(r, binary.LittleEndian, &v)
	case 4:
		var v int32
		return v, binary.Read(r, binary.LittleEndian, &v)
	case 5:
		return readASCII8(r)
	case 6:
		return readASCII8(r)
	case 7:
		assembly, err := readASCII8(r)
		if err != nil {
			return nil, err
		}
		className, err := readASCII8(r)
		if err != nil {
			return nil, err
		}
		jsonText, err := readUTF16LE32(r)
		if err != nil {
			return nil, err
		}
		return JSONClassObject{AssemblyName: assembly, ClassName: className, JSONText: jsonText}, nil
	default:
		return nil, fmt.Errorf("unknown serialized object type %d", objectType)
	}
}

func BundleSize(loc ResourceLocation) (int64, bool) {
	obj, ok := loc.Data.(JSONClassObject)
	if !ok {
		return 0, false
	}
	var payload struct {
		BundleSize int64 `json:"m_BundleSize"`
	}
	if err := json.Unmarshal([]byte(obj.JSONText), &payload); err != nil {
		return 0, false
	}
	if payload.BundleSize <= 0 {
		return 0, false
	}
	return payload.BundleSize, true
}

func readASCII8(r io.Reader) (string, error) {
	var size uint8
	if err := binary.Read(r, binary.LittleEndian, &size); err != nil {
		return "", err
	}
	buf := make([]byte, size)
	if _, err := io.ReadFull(r, buf); err != nil {
		return "", err
	}
	return string(buf), nil
}

func readASCII32(r io.Reader) (string, error) {
	var size int32
	if err := binary.Read(r, binary.LittleEndian, &size); err != nil {
		return "", err
	}
	if size < 0 {
		return "", fmt.Errorf("invalid ascii byte length %d", size)
	}
	buf := make([]byte, size)
	if _, err := io.ReadFull(r, buf); err != nil {
		return "", err
	}
	return string(buf), nil
}

func readUTF16LE32(r io.Reader) (string, error) {
	var size int32
	if err := binary.Read(r, binary.LittleEndian, &size); err != nil {
		return "", err
	}
	if size < 0 {
		return "", fmt.Errorf("invalid utf16 byte length %d", size)
	}
	buf := make([]byte, size)
	if _, err := io.ReadFull(r, buf); err != nil {
		return "", err
	}
	if len(buf)%2 != 0 {
		return "", fmt.Errorf("invalid utf16 byte length %d", len(buf))
	}
	words := make([]uint16, len(buf)/2)
	for i := range words {
		words[i] = binary.LittleEndian.Uint16(buf[i*2:])
	}
	return string(utf16.Decode(words)), nil
}
