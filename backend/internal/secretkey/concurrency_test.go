package secretkey

import (
	"bytes"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func TestConcurrentFirstUseKeepsOneKey(t *testing.T) {
	path := filepath.Join(t.TempDir(), "keys", "master.key")
	keys := make([][]byte, 32)
	var wg sync.WaitGroup
	for i := range keys {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			var err error
			keys[i], err = Load(path)
			if err != nil {
				t.Error(err)
			}
		}(i)
	}
	wg.Wait()
	disk, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range keys {
		if !bytes.Equal(key, disk) {
			t.Fatal("caller received a discarded key")
		}
	}
}
