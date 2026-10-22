package handler

import (
	"auth-api/internal/config"
	"bytes"
	"mime/multipart"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestUploadLimitBeforeMultipartParsing(t *testing.T) {
	f := NewFileService(nil, config.UploadConfig{Root: t.TempDir()})
	for _, declared := range []bool{false, true} {
		var b bytes.Buffer
		mw := multipart.NewWriter(&b)
		part, _ := mw.CreateFormFile("file", "small.txt")
		part.Write([]byte("small"))
		part, _ = mw.CreateFormFile("ignored", "large.bin")
		part.Write(bytes.Repeat([]byte("x"), 2<<20))
		mw.Close()
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest("POST", "/upload", &b)
		c.Request.Header.Set("Content-Type", mw.FormDataContentType())
		if !declared {
			c.Request.ContentLength = -1
		}
		_, _, _, _, _, err := f.Save(c, "avatars", 1024, false)
		if err == nil {
			t.Fatal("oversized ignored multipart part accepted")
		}
		if c.Request.MultipartForm != nil {
			c.Request.MultipartForm.RemoveAll()
		}
	}
}

func TestImageDownloadsRejectDirectorySymlinkAndActiveContent(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "avatars"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(outside, "secret.png"), []byte("secret"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "avatars", "active.svg"), []byte(`<svg xmlns="http://www.w3.org/2000/svg"><script>alert(1)</script></svg>`), 0600); err != nil {
		t.Fatal(err)
	}
	f := NewFileService(nil, config.UploadConfig{Root: root})
	g := gin.New()
	f.MountDownloads(g)
	paths := []string{"/uploads/avatars/", "/uploads/avatars/active.svg"}
	if err := os.Symlink(filepath.Join(outside, "secret.png"), filepath.Join(root, "avatars", "escape.png")); err == nil {
		paths = append(paths, "/uploads/avatars/escape.png")
	} else {
		t.Log("symlink unavailable:", err)
	}
	for _, path := range paths {
		w := httptest.NewRecorder()
		g.ServeHTTP(w, httptest.NewRequest("GET", path, nil))
		if w.Code != 404 {
			t.Fatalf("unsafe image %s => %d", path, w.Code)
		}
	}
}
