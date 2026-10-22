package auth

import (
	"context"
	"crypto/rand"
	"github.com/mojocn/base64Captcha"
	"github.com/redis/go-redis/v9"
	"math/big"
	"strings"
	"time"
)

func NewCaptchaService(r *redis.Client) *CaptchaService { return &CaptchaService{rdb: r} }

func (s *CaptchaService) Generate(ctx context.Context) (*GenerateResult, error) {
	n, e := rand.Int(rand.Reader, big.NewInt(4))
	if e != nil {
		return nil, e
	}
	kind := []string{"string", "math", "chinese", "click"}[n.Int64()]
	if kind == "click" {
		return s.generateClick(ctx)
	}
	var d base64Captcha.Driver
	switch kind {
	case "math":
		d = base64Captcha.NewDriverMath(60, 180, 3, 2, nil, nil, nil)
	case "chinese":
		d = base64Captcha.NewDriverChinese(60, 180, 2, 2, 3, "天地山水风云花草春夏秋冬日月星海", nil, nil, []string{"wqy-microhei.ttc"})
	default:
		d = base64Captcha.NewDriverString(60, 180, 4, 2, 4, "abcdefghjkmnpqrstuvwxyz23456789", nil, nil, nil)
	}
	id, content, answer := d.GenerateIdQuestionAnswer()
	item, e := d.DrawCaptcha(content)
	if e != nil {
		return nil, e
	}
	if e = s.rdb.Set(ctx, "captcha:"+id, strings.ToLower(answer), 5*time.Minute).Err(); e != nil {
		return nil, e
	}
	return &GenerateResult{ID: id, Image: item.EncodeB64string(), Type: kind}, nil
}
func (s *CaptchaService) Verify(ctx context.Context, id, answer string) bool {
	if id == "" || answer == "" {
		return false
	}
	v, e := s.rdb.GetDel(ctx, "captcha:"+id).Result()
	if e == nil && strings.HasPrefix(v, "click:") {
		return verifyClick(strings.TrimPrefix(v, "click:"), answer)
	}
	return e == nil && strings.EqualFold(v, strings.TrimSpace(answer))
}
func (s *CaptchaService) NeedCaptcha(ctx context.Context, ip string) bool {
	v, e := s.rdb.Get(ctx, "adminfail:"+ip).Int()
	return e != nil && e != redis.Nil || v >= 3
}
func (s *CaptchaService) RecordFailure(ctx context.Context, ip string) error {
	script := redis.NewScript(`local n=redis.call('INCR',KEYS[1]); if n==1 then redis.call('EXPIRE',KEYS[1],900) end; return n`)
	return script.Run(ctx, s.rdb, []string{"adminfail:" + ip}).Err()
}
func (s *CaptchaService) ResetFailure(ctx context.Context, ip string) error {
	return s.rdb.Del(ctx, "adminfail:"+ip).Err()
}
