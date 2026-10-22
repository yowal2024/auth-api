package handler

import "testing"

func TestAccountValidationAndSpreadsheetEscaping(t *testing.T) {
	for _, raw := range []string{`oops`, `{}`, `["1.2.3.4"]`, `[{"ip":"invalid"}]`, `[{"ip":"1.2.3.4","lat":91}]`} {
		if validateIPList(raw) == nil {
			t.Errorf("accepted bad region data: %s", raw)
		}
	}
	for _, raw := range []string{"", `[]`, `[{"ip":"1.2.3.4","lat":30,"lng":120,"time":10}]`} {
		if err := validateIPList(raw); err != nil {
			t.Error(err)
		}
	}
	negative, zero := -1, 0
	if validateAccountLimits(&negative, nil) == nil || validateAccountLimits(nil, &negative) == nil || validateAccountLimits(&zero, &zero) != nil {
		t.Fatal("limit validation")
	}
	for _, value := range []string{"=1+1", "  =1+1", "\t=1+1", "\r+1", "\n@SUM(1)", "+1", "-1"} {
		if spreadsheetSafe(value) != "'"+value {
			t.Errorf("unsafe export %q", value)
		}
	}
	if spreadsheetSafe("普通昵称") != "普通昵称" {
		t.Fatal("plain name changed")
	}
}
