package auth

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"
)

func TestManagedSessionsIsolationRevocationAndHeartbeat(t *testing.T) {
	_, r := redisTest(t)
	ctx := context.Background()
	s := NewUserSessionStore(r, time.Hour)
	first, e := s.CreateWithTTL(ctx, &UserSession{UID: 7, AppID: 1, LastIP: "1.2.3.4"}, time.Hour)
	if e != nil {
		t.Fatal(e)
	}
	second, e := s.CreateWithTTL(ctx, &UserSession{UID: 7, AppID: 1}, time.Hour)
	if e != nil {
		t.Fatal(e)
	}
	card, e := s.CreateWithTTL(ctx, &UserSession{UID: 7, AppID: 1, IsKami: true}, time.Hour)
	if e != nil {
		t.Fatal(e)
	}
	if e = s.Touch(ctx, first, "8.8.8.8", true); e != nil {
		t.Fatal(e)
	}
	v, e := s.List(ctx, 1, 7, false)
	if e != nil || len(v) != 2 {
		t.Fatalf("%+v %v", v, e)
	}
	for _, row := range v {
		if row.ID == sessionID(first) && (row.LastHeartbeatAt == 0 || row.LastIP != "8.8.8.8" || row.LoginIP != "1.2.3.4") {
			t.Fatal(row)
		}
		if row.ID == first || row.ID == second {
			t.Fatal("token exposed")
		}
	}
	if _, e = s.Revoke(ctx, 2, 7, false, sessionID(first)); !errors.Is(e, ErrSession) {
		t.Fatal("cross-app revoke", e)
	}
	if _, e = s.Revoke(ctx, 1, 7, false, sessionID(first)); e != nil {
		t.Fatal(e)
	}
	if _, e = s.Get(ctx, first); !errors.Is(e, ErrSession) {
		t.Fatal(e)
	}
	if e = s.Touch(ctx, first, "", true); !errors.Is(e, ErrSession) {
		t.Fatal("touch resurrected revoked token", e)
	}
	// A pre-upgrade unindexed token must also be invalidated by revoke-all.
	raw, _ := json.Marshal(UserSession{UID: 7, AppID: 1, TTLSec: 3600})
	r.Set(ctx, "usess:legacy", raw, time.Hour)
	if e = s.RevokeAll(ctx, 1, 7, false); e != nil {
		t.Fatal(e)
	}
	for _, token := range []string{second, "legacy"} {
		if _, e = s.Get(ctx, token); !errors.Is(e, ErrSession) {
			t.Fatal("revoked session accepted", e)
		}
	}
	if _, e = s.Get(ctx, card); e != nil {
		t.Fatal("card account incorrectly revoked", e)
	}
	fresh, e := s.CreateWithTTL(ctx, &UserSession{UID: 7, AppID: 1}, time.Hour)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = s.Get(ctx, fresh); e != nil {
		t.Fatal("cannot log back in", e)
	}
}

func TestDeviceRevocationPreservesOtherDevicesAndAccounts(t *testing.T) {
	_, r := redisTest(t)
	ctx := context.Background()
	s := NewUserSessionStore(r, time.Hour)
	for _, kami := range []bool{false, true} {
		var tokens []string
		for _, v := range []UserSession{
			{UID: 7, AppID: 1, IsKami: kami, DeviceID: 11},
			{UID: 7, AppID: 1, IsKami: kami, DeviceID: 11},
			{UID: 7, AppID: 1, IsKami: kami, DeviceID: 12},
			{UID: 8, AppID: 1, IsKami: kami, DeviceID: 11},
			{UID: 7, AppID: 2, IsKami: kami, DeviceID: 11},
			{UID: 7, AppID: 1, IsKami: !kami, DeviceID: 11},
		} {
			token, err := s.CreateWithTTL(ctx, &v, time.Hour)
			if err != nil {
				t.Fatal(err)
			}
			tokens = append(tokens, token)
		}
		if err := s.RevokeDevice(ctx, 1, 7, kami, 11); err != nil {
			t.Fatal(err)
		}
		for i, token := range tokens {
			_, err := s.Get(ctx, token)
			if i < 2 && !errors.Is(err, ErrSession) {
				t.Fatal("unbound device token survived", i, err)
			}
			if i >= 2 && err != nil {
				t.Fatal("unrelated session revoked", i, err)
			}
		}
		rows, err := s.List(ctx, 1, 7, kami)
		if err != nil {
			t.Fatal(err)
		}
		for _, row := range rows {
			if row.ID == sessionID(tokens[0]) || row.ID == sessionID(tokens[1]) {
				t.Fatal("revoked device still listed")
			}
		}
	}
}
