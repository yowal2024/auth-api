package geo

import (
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/oschwald/maxminddb-golang"
)

const LocalProvider = "dbip"
const maxDatabaseBytes = 256 << 20
const maxDownloadBytes = 128 << 20
const Attribution = "IP Geolocation by DB-IP (CC BY 4.0)"

var ErrUpdateBusy = errors.New("本地 IP 库正在更新，请稍后重试")

// LocalDB keeps the last validated reader available while downloading a replacement.
// The file lives outside uploads and survives restarts in the runtime data volume.
type LocalDB struct {
	mu         sync.RWMutex
	update     sync.Mutex
	reader     *maxminddb.Reader
	path       string
	updatedAt  int64
	lastError  string
	updating   bool
	stage      string
	downloaded int64
	total      int64
	client     *http.Client
}

type LocalStatus struct {
	Ready           bool   `json:"ready"`
	Updating        bool   `json:"updating"`
	BuildTime       uint   `json:"buildTime"`
	UpdatedAt       int64  `json:"updatedAt"`
	DatabaseType    string `json:"databaseType"`
	LastError       string `json:"lastError"`
	Attribution     string `json:"attribution"`
	Stage           string `json:"stage"`
	DownloadedBytes int64  `json:"downloadedBytes"`
	TotalBytes      int64  `json:"totalBytes"`
}

func NewLocalDB(dir string) *LocalDB {
	l := &LocalDB{path: filepath.Join(dir, "dbip-city-lite.mmdb"), client: &http.Client{
		Timeout:       3 * time.Minute,
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error { return http.ErrUseLastResponse },
	}}
	r, err := readDatabase(l.path)
	if err == nil {
		l.reader = r
		if st, e := os.Stat(l.path); e == nil {
			l.updatedAt = st.ModTime().Unix()
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		l.lastError = "本地 IP 库无法读取，请重新更新"
	}
	return l
}

func (l *LocalDB) Status() LocalStatus {
	if l == nil {
		return LocalStatus{Attribution: Attribution}
	}
	l.mu.RLock()
	defer l.mu.RUnlock()
	s := LocalStatus{Ready: l.reader != nil, Updating: l.updating, UpdatedAt: l.updatedAt, LastError: l.lastError, Attribution: Attribution}
	s.Stage, s.DownloadedBytes, s.TotalBytes = l.stage, l.downloaded, l.total
	if l.reader != nil {
		s.BuildTime = l.reader.Metadata.BuildEpoch
		s.DatabaseType = l.reader.Metadata.DatabaseType
	}
	return s
}

func (l *LocalDB) Close() error {
	if l == nil {
		return nil
	}
	l.update.Lock()
	defer l.update.Unlock()
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.reader != nil {
		err := l.reader.Close()
		l.reader = nil
		return err
	}
	return nil
}

type localNames struct {
	Names map[string]string `maxminddb:"names"`
	ISO   string            `maxminddb:"iso_code"`
}
type localRecord struct {
	Country      localNames   `maxminddb:"country"`
	City         localNames   `maxminddb:"city"`
	Subdivisions []localNames `maxminddb:"subdivisions"`
	Location     struct {
		Latitude  *float64 `maxminddb:"latitude"`
		Longitude *float64 `maxminddb:"longitude"`
	} `maxminddb:"location"`
}

func localName(n localNames) string {
	for _, lang := range []string{"zh-CN", "en"} {
		if n.Names[lang] != "" {
			return n.Names[lang]
		}
	}
	return n.ISO
}
func (l *LocalDB) Lookup(ip string) (Point, error) {
	a := net.ParseIP(ip)
	if a == nil || !a.IsGlobalUnicast() || a.IsPrivate() {
		return Point{}, errors.New("请输入可定位的公网 IPv4 或 IPv6 地址")
	}
	if l == nil {
		return Point{}, errors.New("本地 IP 库尚未安装，请先更新")
	}
	l.mu.RLock()
	defer l.mu.RUnlock()
	if l.reader == nil {
		return Point{}, errors.New("本地 IP 库尚未安装，请先更新")
	}
	var record localRecord
	if err := l.reader.Lookup(a, &record); err != nil {
		return Point{}, errors.New("本地 IP 库查询失败")
	}
	return record.point()
}
func (r localRecord) point() (Point, error) {
	lat, lng := r.Location.Latitude, r.Location.Longitude
	if lat == nil || lng == nil || math.IsNaN(*lat) || math.IsNaN(*lng) || math.IsInf(*lat, 0) || math.IsInf(*lng, 0) || *lat < -90 || *lat > 90 || *lng < -180 || *lng > 180 {
		return Point{}, errors.New("本地 IP 库未收录此 IP 的有效经纬度")
	}
	p := Point{Lat: *lat, Lng: *lng, Nation: localName(r.Country), City: localName(r.City)}
	if len(r.Subdivisions) > 0 {
		p.Province = localName(r.Subdivisions[0])
	}
	return p, nil
}

func readDatabase(path string) (*maxminddb.Reader, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	b, err := io.ReadAll(io.LimitReader(f, maxDatabaseBytes+1))
	if err != nil {
		return nil, err
	}
	if len(b) > maxDatabaseBytes {
		return nil, errors.New("IP 库文件过大")
	}
	r, err := maxminddb.FromBytes(b)
	if err != nil {
		return nil, errors.New("IP 库文件格式无效")
	}
	if !strings.EqualFold(r.Metadata.DatabaseType, "DBIP-City-Lite") || r.Metadata.IPVersion != 6 || r.Metadata.BuildEpoch == 0 {
		r.Close()
		return nil, errors.New("需要 DB-IP City Lite IPv4/IPv6 城市库")
	}
	if err = r.Verify(); err != nil {
		r.Close()
		return nil, fmt.Errorf("IP 库完整性校验失败: %w", err)
	}
	return r, nil
}

// Update only contacts the official HTTPS host; no user supplied URLs or redirects.
// At a month boundary an unpublished current release falls back to last month.
func (l *LocalDB) Update(ctx context.Context) (status LocalStatus, err error) {
	if l == nil {
		return status, errors.New("本地 IP 库服务未配置")
	}
	if !l.update.TryLock() {
		return l.Status(), ErrUpdateBusy
	}
	defer l.update.Unlock()
	l.mu.Lock()
	l.updating = true
	l.lastError = ""
	l.stage, l.downloaded, l.total = "connecting", 0, 0
	l.mu.Unlock()
	defer func() {
		l.mu.Lock()
		l.updating = false
		if err != nil {
			l.lastError = err.Error()
			l.stage = "failed"
		} else {
			l.stage = "complete"
		}
		l.mu.Unlock()
		status = l.Status()
	}()
	now := time.Now().UTC()
	month := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	for i := 0; i < 2; i++ {
		date := month.AddDate(0, -i, 0)
		url := "https://download.db-ip.com/free/dbip-city-lite-" + date.Format("2006-01") + ".mmdb.gz"
		var missing bool
		missing, err = l.download(ctx, url)
		if !missing {
			return status, err
		}
	}
	return status, errors.New("官方尚未提供可用的本月或上月 IP 库")
}

func (l *LocalDB) download(ctx context.Context, url string) (bool, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return false, err
	}
	req.Header.Set("User-Agent", "AuthAPI-LocalGeo/1.0")
	resp, err := l.client.Do(req)
	if err != nil {
		return false, errors.New("下载 IP 库失败，请检查服务器网络后重试")
	}
	defer resp.Body.Close()
	if resp.StatusCode == 404 {
		return true, nil
	}
	if resp.StatusCode != 200 {
		return false, fmt.Errorf("IP 库下载服务返回 HTTP %d", resp.StatusCode)
	}
	if resp.ContentLength > maxDownloadBytes {
		return false, errors.New("IP 库压缩文件过大")
	}
	l.mu.Lock()
	l.stage, l.downloaded, l.total = "downloading", 0, max(resp.ContentLength, 0)
	l.mu.Unlock()
	compressed := &io.LimitedReader{R: &downloadProgressReader{reader: resp.Body, db: l}, N: maxDownloadBytes + 1}
	z, err := gzip.NewReader(compressed)
	if err != nil {
		return false, errors.New("IP 库压缩文件无效")
	}
	defer z.Close()
	if err = os.MkdirAll(filepath.Dir(l.path), 0700); err != nil {
		return false, errors.New("无法创建 IP 库目录")
	}
	f, err := os.CreateTemp(filepath.Dir(l.path), ".dbip-*.tmp")
	if err != nil {
		return false, errors.New("无法创建 IP 库临时文件")
	}
	name := f.Name()
	defer os.Remove(name)
	n, copyErr := io.Copy(f, io.LimitReader(z, maxDatabaseBytes+1))
	if copyErr == nil {
		copyErr = f.Sync()
	}
	closeErr := f.Close()
	if copyErr != nil || closeErr != nil || n > maxDatabaseBytes || compressed.N <= 1 {
		return false, errors.New("IP 库下载不完整、过大或写入失败")
	}
	l.mu.Lock()
	l.stage = "verifying"
	l.mu.Unlock()
	r, err := readDatabase(name)
	if err != nil {
		return false, err
	}
	if err = ctx.Err(); err != nil {
		r.Close()
		return false, errors.New("IP 库更新已取消，已保留当前版本")
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.reader != nil && r.Metadata.BuildEpoch < l.reader.Metadata.BuildEpoch {
		r.Close()
		return false, errors.New("下载版本早于当前 IP 库，已保留当前版本")
	}
	if err = os.Rename(name, l.path); err != nil {
		r.Close()
		return false, errors.New("替换 IP 库失败，已保留当前版本")
	}
	old := l.reader
	l.reader = r
	l.updatedAt = time.Now().Unix()
	if old != nil {
		old.Close()
	}
	return false, nil
}

// Count compressed network bytes, matching Content-Length (not inflated MMDB size).
type downloadProgressReader struct {
	reader io.Reader
	db     *LocalDB
}

func (r *downloadProgressReader) Read(p []byte) (int, error) {
	n, err := r.reader.Read(p)
	r.db.mu.Lock()
	r.db.downloaded += int64(n)
	r.db.mu.Unlock()
	return n, err
}
