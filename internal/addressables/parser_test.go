package addressables

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"testing"
)

func TestParseCatalog(t *testing.T) {
	keyPayload := new(bytes.Buffer)
	must(binary.Write(keyPayload, binary.LittleEndian, int32(1)), t)
	keyOffset := int32(keyPayload.Len())
	keyPayload.WriteByte(0)
	must(binary.Write(keyPayload, binary.LittleEndian, int32(len("bundle_001"))), t)
	keyPayload.WriteString("bundle_001")

	bucketPayload := new(bytes.Buffer)
	must(binary.Write(bucketPayload, binary.LittleEndian, int32(1)), t)
	must(binary.Write(bucketPayload, binary.LittleEndian, keyOffset), t)
	must(binary.Write(bucketPayload, binary.LittleEndian, int32(1)), t)
	must(binary.Write(bucketPayload, binary.LittleEndian, int32(0)), t)

	extraPayload := new(bytes.Buffer)
	extraPayload.WriteByte(7)
	extraPayload.WriteByte(uint8(len("assembly")))
	extraPayload.WriteString("assembly")
	extraPayload.WriteByte(uint8(len("class")))
	extraPayload.WriteString("class")
	utf16JSON := utf16le(`{"m_BundleSize":123}`)
	must(binary.Write(extraPayload, binary.LittleEndian, int32(len(utf16JSON))), t)
	extraPayload.Write(utf16JSON)

	entryPayload := new(bytes.Buffer)
	must(binary.Write(entryPayload, binary.LittleEndian, int32(1)), t)
	for _, v := range []int32{0, 0, -1, 0, 0, 0, 0} {
		must(binary.Write(entryPayload, binary.LittleEndian, v), t)
	}

	catalogJSON, err := json.Marshal(map[string]any{
		"m_ProviderIds":      []string{"Ortega.Common.OrtegaAssestBundleProvider"},
		"m_InternalIds":      []string{"internal/bundle_001"},
		"m_KeyDataString":    base64.StdEncoding.EncodeToString(keyPayload.Bytes()),
		"m_BucketDataString": base64.StdEncoding.EncodeToString(bucketPayload.Bytes()),
		"m_EntryDataString":  base64.StdEncoding.EncodeToString(entryPayload.Bytes()),
		"m_ExtraDataString":  base64.StdEncoding.EncodeToString(extraPayload.Bytes()),
	})
	if err != nil {
		t.Fatal(err)
	}

	catalog, err := Parse(catalogJSON)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if len(catalog.Resources) != 1 {
		t.Fatalf("len(Resources) = %d, want 1", len(catalog.Resources))
	}
	loc := catalog.Resources[0]
	if loc.PrimaryKey != "bundle_001" {
		t.Fatalf("PrimaryKey = %q, want bundle_001", loc.PrimaryKey)
	}
	size, ok := BundleSize(loc)
	if !ok || size != 123 {
		t.Fatalf("BundleSize() = (%d, %v), want (123, true)", size, ok)
	}
}

func TestDecodeEntriesRejectsNegativeCount(t *testing.T) {
	payload := new(bytes.Buffer)
	must(binary.Write(payload, binary.LittleEndian, int32(-1)), t)
	if _, err := decodeEntries(base64.StdEncoding.EncodeToString(payload.Bytes())); err == nil {
		t.Fatal("decodeEntries() error = nil, want error")
	}
}

func utf16le(s string) []byte {
	buf := new(bytes.Buffer)
	for _, r := range s {
		must(binary.Write(buf, binary.LittleEndian, uint16(r)), nil)
	}
	return buf.Bytes()
}

func must(err error, t *testing.T) {
	if err != nil {
		if t != nil {
			t.Fatal(err)
		}
		panic(err)
	}
}
