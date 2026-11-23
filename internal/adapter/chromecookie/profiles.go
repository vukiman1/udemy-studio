package chromecookie

import (
	"cmp"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
)

const localStateFile = "Local State"

type localState struct {
	Profile struct {
		InfoCache map[string]struct {
			Name string `json:"name"`
		} `json:"info_cache"`
	} `json:"profile"`
}

type profileEntry struct {
	dir  string
	name string
}

func readProfiles(dataDir string) ([]profileEntry, error) {
	raw, err := os.ReadFile(filepath.Join(dataDir, localStateFile)) //nolint:gosec // reason: dataDir is the configured Chrome data directory
	if err != nil {
		return nil, wrapAccessError(fmt.Errorf("read Chrome %s: %w", localStateFile, err))
	}
	var state localState
	if err := json.Unmarshal(raw, &state); err != nil {
		return nil, fmt.Errorf("decode Chrome %s: %w", localStateFile, err)
	}
	profiles := make([]profileEntry, 0, len(state.Profile.InfoCache))
	for dir, info := range state.Profile.InfoCache {
		profiles = append(profiles, profileEntry{dir: dir, name: info.Name})
	}
	slices.SortFunc(profiles, func(left, right profileEntry) int { return cmp.Compare(left.dir, right.dir) })
	return profiles, nil
}
