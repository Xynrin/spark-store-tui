package preview

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestFetchCachesImage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		_, _ = writer.Write([]byte("image-data"))
	}))
	defer server.Close()

	path := filepath.Join(t.TempDir(), "image.img")
	if err := fetch(server.URL, path); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "image-data" {
		t.Fatalf("image content = %q", data)
	}
}

func TestFetchRejectsOversizeImage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Length", "12582913")
		writer.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	if err := fetch(server.URL, filepath.Join(t.TempDir(), "image.img")); err == nil {
		t.Fatal("expected oversized image error")
	}
}
