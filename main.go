package main

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/andybalholm/brotli"
	_ "github.com/andybalholm/brotli"

	"github.com/Vilsol/go-pob-data/extractor"
)

var filesToExport = []string{
	"Data/PassiveTreeExpansionJewels.dat64",
	"Data/PassiveTreeExpansionSkills.dat64",
	"Data/PassiveTreeExpansionSpecialSkills.dat64",
	"Data/CostTypes.dat64",
	"Data/Mods.dat64",
	"Data/ActiveSkills.dat64",
	"Data/Essences.dat64",
	"Data/CraftingBenchOptions.dat64",
	"Data/PantheonPanelLayout.dat64",
	"Data/WeaponTypes.dat64",
	"Data/ArmourTypes.dat64",
	"Data/ShieldTypes.dat64",
	"Data/Flasks.dat64",
	"Data/ComponentCharges.dat64",
	"Data/ComponentAttributeRequirements.dat64",
	"Data/BaseItemTypes.dat64",
	"Data/Stats.dat64",
	"Data/AlternatePassiveSkills.dat64",
	"Data/AlternatePassiveAdditions.dat64",
	"Data/DefaultMonsterStats.dat64",
	"Data/SkillTotemVariations.dat64",
	"Data/MonsterVarieties.dat64",
	"Data/MonsterMapDifficulty.dat64",
	"Data/MonsterMapBossDifficulty.dat64",
	"Data/GrantedEffects.dat64",
	"Data/SkillTotems.dat64",
	"Data/GrantedEffectStatSetsPerLevel.dat64",
	"Data/GrantedEffectsPerLevel.dat64",
	"Data/GrantedEffectQualityStats.dat64",
	"Data/SkillGems.dat64",
	"Data/ItemExperiencePerLevel.dat64",
	"Data/Tags.dat64",
	"Data/ActiveSkillType.dat64",
	"Data/ItemClasses.dat64",
	"Data/GrantedEffectStatSets.dat64",
}

const GGG_REPO_BASE = "https://raw.githubusercontent.com/grindinggear/skilltree-export/%s/"

var skillTreeSpriteGroups = []string{
	"background",
	"normalActive",
	"notableActive",
	"keystoneActive",
	"normalInactive",
	"notableInactive",
	"keystoneInactive",
	"mastery",
	"masteryConnected",
	"masteryActiveSelected",
	"masteryInactive",
	"masteryActiveEffect",
	"ascendancyBackground",
	"ascendancy",
	"startNode",
	"groupBackground",
	"frame",
	"jewel",
	"line",
	"jewelRadius",
}

func main() {
	if len(os.Args) < 2 {
		println("please provide path to the game directory")
		os.Exit(1)
		return
	}

	if len(os.Args) < 3 {
		println("please provide passive tree version")
		os.Exit(1)
		return
	}

	if len(os.Args) < 4 {
		println("please provide game version")
		os.Exit(1)
		return
	}

	gamePath := os.Args[1]
	treeVersion := os.Args[2]
	gameVersion := os.Args[3]

	extractRawData(gamePath, gameVersion)
	downloadTreeData(treeVersion, gameVersion)
}

func extractRawData(gamePath string, gameVersion string) {
	if _, err := os.Stat(filepath.Join(gamePath, "Bundles2", "_.index.bin")); err != nil {
		println(err.Error())
		os.Exit(1)
		return
	}

	extractor.LoadParser()
	loader, err := extractor.GetBundleLoader(os.DirFS(gamePath))
	if err != nil {
		println(err.Error())
		os.Exit(1)
		return
	}

	for _, file := range filesToExport {
		println("Extracting", file)

		data, err := loader.Open(file)
		if err != nil {
			println(err.Error())
			os.Exit(1)
			return
		}

		dat, err := extractor.ParseDat(data, filepath.Base(file))
		if err != nil {
			println(err.Error())
			os.Exit(1)
			return
		}

		// Ensure that all slices are always sorted by key
		sort.Slice(dat, func(i, j int) bool {
			return reflect.ValueOf(dat[i]).Field(0).Int() < reflect.ValueOf(dat[j]).Field(0).Int()
		})

		b, err := json.Marshal(dat)
		if err != nil {
			println(err.Error())
			os.Exit(1)
			return
		}

		outNameGzip := strings.Split(filepath.Base(file), ".")[0] + ".json.gz"
		outPathGzip := filepath.Join("data", gameVersion, "raw", outNameGzip)

		if err := os.MkdirAll(filepath.Dir(outPathGzip), 0755); err != nil {
			if !os.IsExist(err) {
				println(err.Error())
				os.Exit(1)
				return
			}
		}

		fGzip, err := os.OpenFile(outPathGzip, os.O_WRONLY|os.O_CREATE, 0755)
		if err != nil {
			println(err.Error())
			os.Exit(1)
			return
		}

		writerGzip := gzip.NewWriter(fGzip)
		if _, err := writerGzip.Write(b); err != nil {
			println(err.Error())
			os.Exit(1)
			return
		}

		_ = writerGzip.Close()
		_ = fGzip.Close()

		outNameBrotli := strings.Split(filepath.Base(file), ".")[0] + ".json.br"
		outPathBrotli := filepath.Join("data", gameVersion, "raw", outNameBrotli)

		fBrotli, err := os.OpenFile(outPathBrotli, os.O_WRONLY|os.O_CREATE, 0755)
		if err != nil {
			println(err.Error())
			os.Exit(1)
			return
		}

		writerBrotli := brotli.NewWriter(fBrotli)
		if _, err := writerBrotli.Write(b); err != nil {
			println(err.Error())
			os.Exit(1)
			return
		}

		_ = writerBrotli.Close()
		_ = fBrotli.Close()
	}
}

func downloadTreeData(treeVersion string, gameVersion string) {
	repoVersionBase := fmt.Sprintf(GGG_REPO_BASE, treeVersion)
	response, err := http.DefaultClient.Get(repoVersionBase + "/data.json")
	if err != nil {
		println(err.Error())
		os.Exit(1)
		return
	}

	defer response.Body.Close()
	dataJson, err := io.ReadAll(response.Body)
	if err != nil {
		println(err.Error())
		os.Exit(1)
		return
	}

	dataJsonOutPathGzip := filepath.Join("data", gameVersion, "tree", "data.json.gz")

	if err := os.MkdirAll(filepath.Dir(dataJsonOutPathGzip), 0755); err != nil {
		if !os.IsExist(err) {
			println(err.Error())
			os.Exit(1)
			return
		}
	}

	fGzip, err := os.OpenFile(dataJsonOutPathGzip, os.O_WRONLY|os.O_CREATE, 0755)
	if err != nil {
		println(err.Error())
		os.Exit(1)
		return
	}

	writerGzip := gzip.NewWriter(fGzip)
	if _, err := writerGzip.Write(dataJson); err != nil {
		println(err.Error())
		os.Exit(1)
		return
	}

	_ = writerGzip.Close()
	_ = fGzip.Close()

	dataJsonOutPathBrotli := filepath.Join("data", gameVersion, "tree", "data.json.br")

	fBrotli, err := os.OpenFile(dataJsonOutPathBrotli, os.O_WRONLY|os.O_CREATE, 0755)
	if err != nil {
		println(err.Error())
		os.Exit(1)
		return
	}

	writerBrotli := brotli.NewWriter(fBrotli)
	if _, err := writerBrotli.Write(dataJson); err != nil {
		println(err.Error())
		os.Exit(1)
		return
	}

	_ = writerBrotli.Close()
	_ = fBrotli.Close()

	var skillTreeData SkillTreeData
	if err := json.Unmarshal(dataJson, &skillTreeData); err != nil {
		println(err.Error())
		os.Exit(1)
		return
	}

	downloaded := make(map[string]bool)

	selectedResolution := skillTreeData.ImageZoomLevels[len(skillTreeData.ImageZoomLevels)-1]
	for _, group := range skillTreeSpriteGroups {
		spriteGroup := skillTreeData.Sprites[group]
		assetPath := spriteGroup[strconv.FormatFloat(selectedResolution, 'f', -1, 64)]
		if assetPath.Filename == "" {
			for _, p := range spriteGroup {
				assetPath = p
				break
			}
		}

		parsedUrl, err := url.Parse(assetPath.Filename)
		if err != nil {
			println(err.Error())
			os.Exit(1)
			return
		}

		fileName := path.Base(parsedUrl.Path)
		if _, ok := downloaded[fileName]; ok {
			continue
		}
		downloaded[fileName] = true

		println("Downloading", fileName)

		response, err := http.DefaultClient.Get(repoVersionBase + "/assets/" + fileName)
		if err != nil {
			println(err.Error())
			os.Exit(1)
			return
		}

		defer response.Body.Close()
		fileData, err := io.ReadAll(response.Body)
		if err != nil {
			println(err.Error())
			os.Exit(1)
			return
		}

		fileOutPath := filepath.Join("data", gameVersion, "tree", "assets", fileName)
		if err := os.MkdirAll(filepath.Dir(fileOutPath), 0755); err != nil {
			if !os.IsExist(err) {
				println(err.Error())
				os.Exit(1)
				return
			}
		}

		if err := os.WriteFile(fileOutPath, fileData, 0755); err != nil {
			println(err.Error())
			os.Exit(1)
			return
		}
	}
}
