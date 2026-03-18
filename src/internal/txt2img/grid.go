package txt2img

import (
	"fmt"
	"log"
	"mea_go/src/internal"
	"os"
	"slices"
	"strings"
)

type GridSize struct {
	x int
	y int
}

var g4x4 = GridSize{
	x: 4,
	y: 4,
}

type ImgId = string
type SpotId = string

type OtherState struct {
	imageIds   []ImgId
	imageData  map[ImgId]ImgData
	imageSpots map[ImgId]SpotId
	spotHolder map[SpotId]ImgId
}

const emptySpot = ""

func loadOtherState(logger *log.Logger) *OtherState {
	var loadeImgsNum int = 0
	imgDir := internal.DirImage()

	var oStat = OtherState{
		imageIds:   make([]string, 0, 128),
		imageData:  make(map[ImgId]ImgData, 128),
		spotHolder: make(map[string]string, 16),
		imageSpots: make(map[string]string, 16),
	}

	entries, err := os.ReadDir(imgDir)
	if err != nil {
		log.Fatalln("scaning imgs", err.Error())
	}
	panicker := internal.Panicker(4)
	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasSuffix(name, ".png") {
			continue
		}
		basename := strings.TrimSuffix(name, ".png")
		yamlFile := internal.Filename(basename, "yaml")
		_ = yamlFile

		imgPath := strings.Join([]string{imgDir, name}, "/")
		metadataPath := strings.Join([]string{imgDir, name}, "/")

		var xd = ImgMetadata{prompts: []string{"", "", ""}}
		if _, err := os.Stat(metadataPath); err != nil {

		}

		data, err := os.ReadFile(imgPath)
		if panicker.HasError(err) {
			continue
		}

		loadeImgsNum += 1
		oStat.imageIds = append(oStat.imageIds, basename)
		oStat.imageData[basename] = ImgData{
			bytes: data,
			meta:  xd,
		}
	}

	logger.Printf("slots empty prefil")

	var imgCount = 0
	for y := range g4x4.y {
		for x := range g4x4.x {
			spotName := spotName(x, y)
			index := spotIdx(x, y)
			if index >= len(oStat.imageIds) {
				oStat.spotHolder[spotName] = emptySpot
				continue
			}

			var imgId = oStat.imageIds[index]
			logger.Printf("id is %s", imgId)

			oStat.imageSpots[imgId] = spotName
			oStat.spotHolder[spotName] = imgId
			imgCount++
		}
	}
	logger.Printf("actuall content fill: %d imgs places", imgCount)

	return &oStat
}

func (oth *OtherState) deleteImg(imgId ImgId) {
	if spotId, ok := oth.imageSpots[imgId]; ok {
		if _, ok := oth.spotHolder[spotId]; ok {

			oth.spotHolder[spotId] = emptySpot
		}
		delete(oth.imageSpots, imgId)
	}

	if _, ok := oth.imageSpots[imgId]; !ok {
		fmt.Printf("right, img is not present there:D\n")
	}

	delete(oth.imageData, imgId)
	if idx, ok := slices.BinarySearch(oth.imageIds, imgId); ok {
		oth.imageIds[idx] = "deleted"
		fmt.Printf("marked as deleted\n")
	}

}
func (oth *OtherState) addImg(id string, imgBts []byte) {
	oth.imageIds = append(oth.imageIds, id)
	oth.imageData[id] = ImgData{
		meta:  emptyMetadata,
		bytes: imgBts,
	}
}
func (oth *OtherState) placeInNewSpot(id string) SpotId {
	// TODO: find unused meaby
	return spotName(0, 0)
}

func spotName(x int, y int) SpotId {
	return fmt.Sprintf("img_slot_%d_%d", x, y)
}
func spotIdx(x int, y int) int {
	return y*4 + x
}
