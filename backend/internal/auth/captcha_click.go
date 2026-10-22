package auth

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/golang/freetype/truetype"
	"github.com/mojocn/base64Captcha"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math/big"
	"strconv"
	"strings"
	"time"
)

type clickPoint struct{ X, Y int }

func secureInt(max int) (int, error) {
	v, e := rand.Int(rand.Reader, big.NewInt(int64(max)))
	if e != nil {
		return 0, e
	}
	return int(v.Int64()), nil
}
func (s *CaptchaService) generateClick(ctx context.Context) (*GenerateResult, error) {
	id, e := RandomToken()
	if e != nil {
		return nil, e
	}
	chars := []rune("天地山水风云花草春夏秋冬日月星海")
	for i := len(chars) - 1; i > 0; i-- {
		j, e := secureInt(i + 1)
		if e != nil {
			return nil, e
		}
		chars[i], chars[j] = chars[j], chars[i]
	}
	img := image.NewRGBA(image.Rect(0, 0, 300, 150))
	draw.Draw(img, img.Bounds(), image.NewUniform(color.RGBA{245, 248, 250, 255}), image.Point{}, draw.Src)
	face := truetype.NewFace(base64Captcha.DefaultEmbeddedFonts.LoadFontByName("fonts/wqy-microhei.ttc"), &truetype.Options{Size: 26})
	defer face.Close()
	points := []clickPoint{}
	positions := []int{0, 1, 2, 3, 4, 5}
	for i := 5; i > 0; i-- {
		j, e := secureInt(i + 1)
		if e != nil {
			return nil, e
		}
		positions[i], positions[j] = positions[j], positions[i]
	}
	for i := 0; i < 6; i++ {
		jitterX, e := secureInt(20)
		if e != nil {
			return nil, e
		}
		jitterY, e := secureInt(20)
		if e != nil {
			return nil, e
		}
		cell := positions[i]
		x := 25 + (cell%3)*95 + jitterX
		y := 35 + (cell/3)*65 + jitterY
		drawer := font.Drawer{Dst: img, Src: image.NewUniform(color.RGBA{uint8(30 + i*20), uint8(60 + i*9), 100, 255}), Face: face, Dot: fixed.P(x, y)}
		drawer.DrawString(string(chars[i]))
		if i < 3 {
			points = append(points, clickPoint{x + 13, y - 13})
		}
	}
	encoded, _ := json.Marshal(points)
	if e = s.rdb.Set(ctx, "captcha:"+id, "click:"+string(encoded), 5*time.Minute).Err(); e != nil {
		return nil, e
	}
	var out bytes.Buffer
	if e = png.Encode(&out, img); e != nil {
		return nil, e
	}
	return &GenerateResult{ID: id, Type: "click", Prompt: fmt.Sprintf("请依次点击：%c、%c、%c", chars[0], chars[1], chars[2]), Image: "data:image/png;base64," + base64.StdEncoding.EncodeToString(out.Bytes())}, nil
}
func verifyClick(saved, answer string) bool {
	var points []clickPoint
	if json.Unmarshal([]byte(saved), &points) != nil || len(points) != 3 {
		return false
	}
	parts := strings.Split(answer, ";")
	if len(parts) != 3 {
		return false
	}
	for i, v := range parts {
		xy := strings.Split(v, ",")
		if len(xy) != 2 {
			return false
		}
		x, e := strconv.Atoi(xy[0])
		if e != nil {
			return false
		}
		y, e := strconv.Atoi(xy[1])
		if e != nil {
			return false
		}
		dx, dy := x-points[i].X, y-points[i].Y
		if dx < -20 || dx > 20 || dy < -20 || dy > 20 {
			return false
		}
	}
	return true
}
