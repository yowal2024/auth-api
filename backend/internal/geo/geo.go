package geo

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

func Distance(a, b Point) float64 {
	const radius = 6371
	rad := math.Pi / 180
	dlat := (b.Lat - a.Lat) * rad
	dlng := (b.Lng - a.Lng) * rad
	x := math.Sin(dlat/2)*math.Sin(dlat/2) + math.Cos(a.Lat*rad)*math.Cos(b.Lat*rad)*math.Sin(dlng/2)*math.Sin(dlng/2)
	if x > 1 {
		x = 1
	}
	return 2 * radius * math.Atan2(math.Sqrt(x), math.Sqrt(1-x))
}
func Lookup(ctx context.Context, provider, key, ip string) (Point, error) {
	var point Point
	if key == "" {
		return point, fmt.Errorf("地域服务 key 未配置")
	}
	endpoint := ""
	switch provider {
	case "tencent", "":
		endpoint = "https://apis.map.qq.com/ws/location/v1/ip"
	case "amap":
		endpoint = "https://restapi.amap.com/v3/ip"
	default:
		return point, fmt.Errorf("未知地域服务")
	}
	req, e := http.NewRequestWithContext(ctx, "GET", endpoint+"?"+url.Values{"key": {key}, "ip": {ip}}.Encode(), nil)
	if e != nil {
		return point, e
	}
	resp, e := (&http.Client{Timeout: 5 * time.Second}).Do(req)
	if e != nil {
		return point, e
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return point, fmt.Errorf("地域服务 HTTP %d", resp.StatusCode)
	}
	raw, e := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if e != nil {
		return point, e
	}
	if provider != "amap" {
		var v tencentResp
		if e = json.Unmarshal(raw, &v); e != nil {
			return point, e
		}
		if v.Status != 0 {
			return point, fmt.Errorf("地域查询失败: %s", v.Message)
		}
		return Point{Lat: v.Result.Location.Lat, Lng: v.Result.Location.Lng, Nation: v.Result.AdInfo.Nation, Province: v.Result.AdInfo.Province, City: v.Result.AdInfo.City, District: v.Result.AdInfo.District}, nil
	}
	var v amapResp
	if e = json.Unmarshal(raw, &v); e != nil {
		return point, e
	}
	if v.Status != "1" {
		return point, fmt.Errorf("地域查询失败: %s", v.Info)
	}
	parts := strings.Split(strings.ReplaceAll(v.Rectangle, ";", ","), ",")
	if len(parts) != 4 {
		return point, fmt.Errorf("地域服务没有返回坐标")
	}
	values := make([]float64, 4)
	for i, s := range parts {
		values[i], e = strconv.ParseFloat(s, 64)
		if e != nil {
			return point, e
		}
	}
	return Point{Lat: (values[1] + values[3]) / 2, Lng: (values[0] + values[2]) / 2, Province: v.Province, City: v.City}, nil
}
func Check(entries []IPEntry, point Point, ip string, limit, distance int) ([]IPEntry, error) {
	for i, v := range entries {
		if v.IP == ip || Distance(Point{Lat: v.Lat, Lng: v.Lng}, point) <= float64(distance) {
			entries[i].Time = time.Now().Unix()
			return entries, nil
		}
	}
	if limit > 0 && len(entries) >= limit {
		return nil, fmt.Errorf("IP 地点数超出限制")
	}
	return append(entries, IPEntry{IP: ip, Lat: point.Lat, Lng: point.Lng, Province: point.Province, City: point.City, Time: time.Now().Unix()}), nil
}
