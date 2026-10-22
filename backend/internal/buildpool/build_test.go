package buildpool

import (
	"archive/zip"
	"auth-api/internal/buildpool/bsdiff"
	"auth-api/internal/config"
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestDeltaRoundTripAndMalformed(t *testing.T) {
	for _, v := range []struct{ old, new []byte }{{[]byte("old content"), []byte("new content longer")}, {bytes.Repeat([]byte("1234"), 10000), bytes.Repeat([]byte("5678"), 30000)}} {
		patch, e := bsdiff.Build(v.old, v.new)
		if e != nil {
			t.Fatal(e)
		}
		got, e := bsdiff.Apply(v.old, patch)
		if e != nil || !bytes.Equal(got, v.new) {
			t.Fatalf("roundtrip %v", e)
		}
		if _, e = bsdiff.Apply(v.old, patch[:20]); e == nil {
			t.Fatal("truncation accepted")
		}
	}
}
func TestAndroidAPKCloneAndIndependentSignature(t *testing.T) {
	sdk := os.Getenv("AUTH_RECOVERY_ANDROID_SDK")
	if sdk == "" {
		t.Skip("set AUTH_RECOVERY_ANDROID_SDK for independent Android validation")
	}
	dir := t.TempDir()
	manifest := `<manifest xmlns:android="http://schemas.android.com/apk/res/android" package="com.recovery.fixture" android:versionCode="1" android:versionName="1"><uses-sdk android:minSdkVersion="24" android:targetSdkVersion="35"/><permission android:name="com.recovery.fixture.PERMISSION"/><application android:label=".UnchangedLabel" android:hasCode="false"><activity android:name=".MainActivity"/><service android:name="Worker"/><provider android:name="com.recovery.fixture.lowercaseprovider" android:authorities="com.recovery.fixture.provider"/></application></manifest>`
	m := filepath.Join(dir, "AndroidManifest.xml")
	if e := os.WriteFile(m, []byte(manifest), 0600); e != nil {
		t.Fatal(e)
	}
	apk := filepath.Join(dir, "original.apk")
	aapt := filepath.Join(sdk, "build-tools", "36.0.0", "aapt2.exe")
	framework := filepath.Join(sdk, "platforms", "android-37.0", "android.jar")
	if out, e := exec.Command(aapt, "link", "--manifest", m, "-I", framework, "-o", apk).CombinedOutput(); e != nil {
		t.Fatalf("aapt: %s %v", out, e)
	}
	raw, e := os.ReadFile(apk)
	if e != nil {
		t.Fatal(e)
	}
	// Include stored resources/native-library bytes to independently check alignment.
	zr, e := zip.NewReader(bytes.NewReader(raw), int64(len(raw)))
	if e != nil {
		t.Fatal(e)
	}
	var extra bytes.Buffer
	zw := zip.NewWriter(&extra)
	for _, f := range zr.File {
		if e = zw.Copy(f); e != nil {
			t.Fatal(e)
		}
	}
	for _, name := range []string{"assets/data.bin", "lib/arm64-v8a/libfixture.so"} {
		h := zip.FileHeader{Name: name, Method: zip.Store}
		w, e := zw.CreateHeader(&h)
		if e != nil {
			t.Fatal(e)
		}
		w.Write(bytes.Repeat([]byte{1, 2, 3, 4}, 501))
	}
	zw.Close()
	raw = extra.Bytes()
	original, e := ExtractPackageName(raw)
	if e != nil || original != "com.recovery.fixture" {
		t.Fatalf("package %q %v", original, e)
	}
	store, e := NewDiffStore(config.CloneConfig{CanonicalDir: dir, OutputDir: dir, DiffTTLSec: 60})
	if e != nil {
		t.Fatal(e)
	}
	meta, e := store.GenerateFromBytes(raw, nil)
	if e != nil {
		t.Fatal(e)
	}
	patch, e := os.ReadFile(filepath.Join(dir, meta.DiffName))
	if e != nil {
		t.Fatal(e)
	}
	clone, e := bsdiff.Apply(raw, patch)
	if e != nil {
		t.Fatal(e)
	}
	pkg, e := ExtractPackageName(clone)
	if e != nil || pkg != meta.NewPackage || pkg == original {
		t.Fatalf("cloned package %s %v", pkg, e)
	}
	if sha(clone) != meta.CloneSha256 {
		t.Fatal("clone hash mismatch")
	}
	signed := filepath.Join(dir, "clone.apk")
	os.WriteFile(signed, clone, 0600)
	if out, e := exec.Command(filepath.Join(sdk, "build-tools", "36.0.0", "zipalign.exe"), "-c", "-P", "16", "-v", "4", signed).CombinedOutput(); e != nil {
		t.Fatalf("zipalign: %s %v", out, e)
	} else {
		t.Log(string(out))
	}
	if out, e := exec.Command(aapt, "dump", "xmltree", signed, "--file", "AndroidManifest.xml").CombinedOutput(); e != nil {
		t.Fatalf("xmltree: %s %v", out, e)
	} else {
		for _, want := range []string{"com.recovery.fixture.MainActivity", "com.recovery.fixture.Worker", "com.recovery.fixture.lowercaseprovider", ".UnchangedLabel", meta.NewPackage + ".provider", meta.NewPackage + ".PERMISSION"} {
			if !strings.Contains(string(out), want) {
				t.Fatalf("manifest missing %s: %s", want, out)
			}
		}
	}
	jar := filepath.Join(sdk, "build-tools", "36.0.0", "lib", "apksigner.jar")
	if out, e := exec.Command("java", "-jar", jar, "verify", "--verbose", "--min-sdk-version", "24", signed).CombinedOutput(); e != nil {
		t.Fatalf("apksigner: %s %v", out, e)
	} else {
		t.Log(string(out))
	}
}
