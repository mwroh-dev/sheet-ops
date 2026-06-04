package requestcompiler

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/mwroh/sheet-ops/runtime/pathid"
)

func NewWorkUnitID(slug string, now time.Time) (string, error) {
	slug, err := pathid.Validate("scenario slug", slug)
	if err != nil {
		return "", err
	}
	var suffix [3]byte
	_, _ = rand.Read(suffix[:])
	return slug + "__" + now.UTC().Format("20060102T150405Z") + "__" + hex.EncodeToString(suffix[:]), nil
}
