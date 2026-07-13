package domain

// SourceKind describes the metadata format exposed by a catalog source.
type SourceKind string

const (
	SourceKindSparkMetadata SourceKind = "spark-metadata"
)

// CatalogSource is a configured software catalog, not a package mirror.
// Catalogs can expose multiple mirrors through the Mirrors field.
type CatalogSource struct {
	ID          string
	Name        string
	Kind        SourceKind
	Description string
	Mirrors     []Mirror
}

// Category is a source-neutral grouping used to browse large catalogs.
type Category struct {
	ID   string
	Name string
}

// Mirror is a downloadable endpoint for a source. Region is a routing hint,
// never a trust signal; all downloaded files still require verification.
type Mirror struct {
	ID       string
	Name     string
	BaseURL  string
	Region   string
	Priority int
}

// App is the source-neutral application representation used by the UI.
type App struct {
	ID             string
	Name           string
	Version        string
	PackageFormat  string
	Size           string
	Description    string
	SourceID       string
	Checksum       string
	Category       string
	PackageName    string
	Filename       string
	MetalinkURL    string
	IconURL        string
	ScreenshotURLs []string
	Assets         []ReleaseAsset
}

// ReleaseAsset is the normalized, directly downloadable package artifact.
// Digest is optional because not all upstream release APIs publish one.
type ReleaseAsset struct {
	Name          string
	DownloadURL   string
	SizeBytes     int64
	SHA256        string
	PackageFormat string
}
