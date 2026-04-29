package boi

import "github.com/vmihailenco/msgpack/v5"

const commonRequireClientUpdate = 103

type getDataURIRequest struct {
	CountryCode string `msgpack:"CountryCode"`
	UserID      int64  `msgpack:"UserId"`
}

type dataURIResponse struct {
	AppAssetVersionInfo struct {
		Version string `msgpack:"Version"`
	} `msgpack:"AppAssetVersionInfo"`
	AssetCatalogURIFormat      string `msgpack:"AssetCatalogUriFormat"`
	AssetCatalogFixedURIFormat string `msgpack:"AssetCatalogFixedUriFormat"`
	MasterURIFormat            string `msgpack:"MasterUriFormat"`
	NoticeBannerImageURIFormat string `msgpack:"NoticeBannerImageUriFormat"`
	RawDataURIFormat           string `msgpack:"RawDataUriFormat"`
}

type apiErrorResponse struct {
	ErrorCode int    `msgpack:"ErrorCode"`
	Message   string `msgpack:"Message"`
}

type DataURIInfo struct {
	AppVersion           string
	MasterVersion        string
	AssetVersion         string
	AssetCatalogFixedFmt string
	MasterURIFmt         string
	RawDataURIFmt        string
	CatalogURIFmt        string
	NoticeBannerImageFmt string
	Response             dataURIResponse
}

type MasterBookCatalog struct {
	MasterBookInfoMap map[string]MasterBookInfo `msgpack:"MasterBookInfoMap"`
}

type MasterBookInfo struct {
	Hash string `msgpack:"Hash"`
	Name string `msgpack:"Name"`
	Size int    `msgpack:"Size"`
}

func marshalMessagePack(v any) ([]byte, error) {
	return msgpack.Marshal(v)
}
