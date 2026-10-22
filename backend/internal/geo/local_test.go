package geo

import (
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/maxmind/mmdbwriter"
	"github.com/maxmind/mmdbwriter/mmdbtype"
)

type geoRoundTrip func(*http.Request) (*http.Response, error)

func (f geoRoundTrip) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func fixtureDatabase(t *testing.T, epoch int64) []byte {
	t.Helper()
	w, err := mmdbwriter.New(mmdbwriter.Options{DatabaseType: "DBIP-City-Lite", Description: map[string]string{"en": "Synthetic city test database"}, IPVersion: 6, BuildEpoch: epoch})
	if err != nil {
		t.Fatal(err)
	}
	for _, subnet := range []string{"1.2.3.0/24", "240e:1234::/32"} {
		_, network, _ := net.ParseCIDR(subnet)
		err = w.Insert(network, mmdbtype.Map{
			"country":      mmdbtype.Map{"names": mmdbtype.Map{"en": mmdbtype.String("China"), "zh-CN": mmdbtype.String("中国")}},
			"subdivisions": mmdbtype.Slice{mmdbtype.Map{"names": mmdbtype.Map{"en": mmdbtype.String("Guangdong")}}},
			"city":         mmdbtype.Map{"names": mmdbtype.Map{"en": mmdbtype.String("Shenzhen")}},
			"location":     mmdbtype.Map{"latitude": mmdbtype.Float64(22.54), "longitude": mmdbtype.Float64(114.06)},
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	var b bytes.Buffer
	if _, err = w.WriteTo(&b); err != nil {
		t.Fatal(err)
	}
	return b.Bytes()
}
func zipped(t *testing.T, b []byte) []byte {
	t.Helper()
	var out bytes.Buffer
	z := gzip.NewWriter(&out)
	if _, err := z.Write(b); err != nil {
		t.Fatal(err)
	}
	if err := z.Close(); err != nil {
		t.Fatal(err)
	}
	return out.Bytes()
}
func responseBytes(b []byte) *http.Response {
	return &http.Response{StatusCode: 200, Body: io.NopCloser(bytes.NewReader(b)), ContentLength: int64(len(b))}
}

func TestLocalDatabaseUpdateLookupAndRestart(t *testing.T) {
	dir := t.TempDir()
	local := NewLocalDB(dir)
	defer local.Close()
	if local.Status().Ready {
		t.Fatal("unexpected initial database")
	}
	if _, err := local.Lookup("1.2.3.4"); err == nil {
		t.Fatal("missing DB must fail")
	}
	epoch := time.Now().Unix() - 100
	body := zipped(t, fixtureDatabase(t, epoch))
	local.client.Transport = geoRoundTrip(func(r *http.Request) (*http.Response, error) {
		if r.URL.Scheme != "https" || r.URL.Host != "download.db-ip.com" || !strings.HasPrefix(r.URL.Path, "/free/dbip-city-lite-") {
			t.Fatal("unexpected update origin", r.URL)
		}
		return responseBytes(body), nil
	})
	s, err := local.Update(context.Background())
	if err != nil || !s.Ready || s.Updating {
		t.Fatal(s, err)
	}
	if s.Stage != "complete" || s.DownloadedBytes != int64(len(body)) || s.TotalBytes != int64(len(body)) {
		t.Fatal("progress must count the compressed download", s)
	}
	for _, ip := range []string{"1.2.3.4", "240e:1234::1", "::ffff:1.2.3.4"} {
		p, err := local.Lookup(ip)
		if err != nil || p.Lat != 22.54 || p.Lng != 114.06 || p.Nation != "中国" || p.City != "Shenzhen" {
			t.Fatal(ip, p, err)
		}
	}
	for _, ip := range []string{"", "bad", "127.0.0.1", "192.168.1.1", "::1", "fc00::1", "8.8.8.8"} {
		if _, err := local.Lookup(ip); err == nil {
			t.Fatal("unknown or private IP accepted", ip)
		}
	}
	// Replacing an existing file works on both Windows and Linux without closing live readers first.
	if _, err = local.Update(context.Background()); err != nil {
		t.Fatal(err)
	}
	reopened := NewLocalDB(dir)
	defer reopened.Close()
	if p, err := reopened.Lookup("1.2.3.4"); err != nil || p.City != "Shenzhen" {
		t.Fatal(p, err)
	}
	original, _ := os.ReadFile(local.path)
	for _, bad := range [][]byte{[]byte("not gzip"), zipped(t, []byte("not mmdb")), body[:len(body)-5], zipped(t, fixtureDatabase(t, epoch-100))} {
		local.client.Transport = geoRoundTrip(func(*http.Request) (*http.Response, error) { return responseBytes(bad), nil })
		if _, err = local.Update(context.Background()); err == nil {
			t.Fatal("bad update accepted")
		}
		if !local.Status().Ready || local.Status().LastError == "" {
			t.Fatal(local.Status())
		}
		if local.Status().Stage != "failed" {
			t.Fatal("failed update status missing")
		}
		current, _ := os.ReadFile(local.path)
		if !bytes.Equal(original, current) {
			t.Fatal("old database changed")
		}
		if _, err = local.Lookup("1.2.3.4"); err != nil {
			t.Fatal(err)
		}
	}
	files, _ := filepath.Glob(filepath.Join(dir, ".dbip-*.tmp"))
	if len(files) > 0 {
		t.Fatal("temporary files leaked", files)
	}
}

func TestLocalUpdateFallbackBusyAndHTTPFailures(t *testing.T) {
	l := NewLocalDB(t.TempDir())
	defer l.Close()
	body := zipped(t, fixtureDatabase(t, time.Now().Unix()))
	calls := 0
	l.client.Transport = geoRoundTrip(func(r *http.Request) (*http.Response, error) {
		calls++
		resp := responseBytes(body)
		if calls == 1 {
			resp.StatusCode = 404
		} else {
			now := time.Now().UTC()
			last := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC).AddDate(0, -1, 0)
			if !strings.Contains(r.URL.Path, last.Format("2006-01")) {
				t.Fatal(r.URL.Path)
			}
		}
		return resp, nil
	})
	if _, err := l.Update(context.Background()); err != nil || calls != 2 {
		t.Fatal(err, calls)
	}
	l.update.Lock()
	_, err := l.Update(context.Background())
	l.update.Unlock()
	if !errors.Is(err, ErrUpdateBusy) {
		t.Fatal(err)
	}
	for _, status := range []int{302, 403, 500} {
		l.client.Transport = geoRoundTrip(func(*http.Request) (*http.Response, error) {
			resp := responseBytes(nil)
			resp.StatusCode = status
			return resp, nil
		})
		if _, err := l.Update(context.Background()); err == nil {
			t.Fatal(status)
		}
	}
	l.client.Transport = geoRoundTrip(func(*http.Request) (*http.Response, error) {
		resp := responseBytes(nil)
		resp.ContentLength = maxDownloadBytes + 1
		return resp, nil
	})
	if _, err := l.Update(context.Background()); err == nil {
		t.Fatal("excessive length accepted")
	}
	if err := l.client.CheckRedirect(nil, nil); !errors.Is(err, http.ErrUseLastResponse) {
		t.Fatal("redirects enabled")
	}
}

func TestMissingCoordinatesAreNotZeroCoordinates(t *testing.T) {
	r := localRecord{}
	if _, err := r.point(); err == nil {
		t.Fatal("missing coordinates accepted")
	}
	zero := 0.0
	r.Location.Latitude = &zero
	r.Location.Longitude = &zero
	if _, err := r.point(); err != nil {
		t.Fatal("explicit zero coordinates are valid", err)
	}
	bad := 181.0
	r.Location.Longitude = &bad
	if _, err := r.point(); err == nil {
		t.Fatal("out of range coordinates accepted")
	}
}

func TestOfficialLocalDatabase(t *testing.T) {
	path := os.Getenv("AUTH_RECOVERY_GEO_FIXTURE")
	if path == "" {
		t.Skip("set AUTH_RECOVERY_GEO_FIXTURE to official mmdb.gz")
	}
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	l := NewLocalDB(t.TempDir())
	defer l.Close()
	l.client.Transport = geoRoundTrip(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Body: io.NopCloser(f)}, nil
	})
	if _, err = l.Update(context.Background()); err != nil {
		t.Fatal(err)
	}
	for _, ip := range []string{"114.114.114.114", "223.5.5.5", "119.29.29.29", "180.76.76.76", "2400:3200::1"} {
		p, err := l.Lookup(ip)
		if err != nil {
			t.Fatal(ip, err)
		}
		t.Log(ip, p)
	}
}
