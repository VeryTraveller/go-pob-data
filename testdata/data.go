package testdata

import "strings"

const GameVersion = "3.22.1.1"

func ShortGameVersion() string {
	return strings.Join(strings.Split(GameVersion, ".")[:2], ".")
}
