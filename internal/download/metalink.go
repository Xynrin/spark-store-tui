package download

import (
	"context"
	"encoding/xml"
	"fmt"
	"net/http"
	"slices"
	"strings"
)

// MetalinkAsset is the package and ranked URLs described by a Metalink v3
// document. Spark currently publishes MD5/SHA-1 entries, so no weak checksum
// is treated as a replacement for SHA-256 verification.
type MetalinkAsset struct {
	Filename string
	URLs     []string
}

func ResolveMetalink(ctx context.Context, client *http.Client, url string) (MetalinkAsset, error) {
	if client == nil {
		client = http.DefaultClient
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return MetalinkAsset{}, err
	}
	response, err := client.Do(request)
	if err != nil {
		return MetalinkAsset{}, err
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return MetalinkAsset{}, fmt.Errorf("metalink request returned %s", response.Status)
	}
	var document metalinkDocument
	if err := xml.NewDecoder(response.Body).Decode(&document); err != nil {
		return MetalinkAsset{}, err
	}
	if len(document.Files) == 0 || len(document.Files[0].URLs) == 0 {
		return MetalinkAsset{}, fmt.Errorf("metalink has no downloadable URLs")
	}
	urls := document.Files[0].URLs
	slices.SortStableFunc(urls, func(left, right metalinkURL) int {
		if left.Preference != right.Preference {
			return right.Preference - left.Preference
		}
		return strings.Compare(left.Location, right.Location)
	})
	result := MetalinkAsset{Filename: document.Files[0].Name, URLs: make([]string, 0, len(urls))}
	for _, mirror := range urls {
		if mirror.Value != "" {
			result.URLs = append(result.URLs, strings.TrimSpace(mirror.Value))
		}
	}
	if result.Filename == "" || len(result.URLs) == 0 {
		return MetalinkAsset{}, fmt.Errorf("metalink package data is incomplete")
	}
	return result, nil
}

type metalinkDocument struct {
	Files []metalinkFile `xml:"files>file"`
}

type metalinkFile struct {
	Name string        `xml:"name,attr"`
	URLs []metalinkURL `xml:"resources>url"`
}

type metalinkURL struct {
	Location   string `xml:"location,attr"`
	Preference int    `xml:"preference,attr"`
	Value      string `xml:",chardata"`
}
