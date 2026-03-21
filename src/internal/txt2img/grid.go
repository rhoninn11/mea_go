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

const spotNum = 16

var g4x4 = GridSize{
	x: 4,
	y: 4,
}

type ImgId = string
type SpotId = string

type Stats struct {
	imgNum     int
	imgMemSize int64
	info       string
}

func (s *Stats) inform() string {
	bldr := strings.Builder{}
	bldr.WriteString(fmt.Sprintf("+++ %d images loaded\n", s.imgNum))

	kb := s.imgMemSize / 1024
	mb := float32(kb) / 1024
	if mb < 0 {
		bldr.WriteString(fmt.Sprintf("+++ using %d kb\n", kb))
	} else {
		bldr.WriteString(fmt.Sprintf("+++ using %02.f mb\n", mb))
	}

	return bldr.String()
}

type OtherState struct {
	imageIds   []ImgId
	imageData  map[ImgId]ImgData
	imageSpots map[ImgId]SpotId
	spotHolder map[SpotId]ImgId

	occupied [spotNum]bool
	stats    *Stats
}

const emptySpot = ""

func loadOtherState(logger *log.Logger) *OtherState {
	var panicker = internal.Panicker(4)
	var imgStats = Stats{}

	imgDir := internal.DirImage()

	var oStat = OtherState{
		imageIds:   make([]string, 0, 128),
		imageData:  make(map[ImgId]ImgData, 128),
		spotHolder: make(map[string]string, spotNum),
		imageSpots: make(map[string]string, spotNum),

		stats: &imgStats,
	}

	entries, err := os.ReadDir(imgDir)
	if err != nil {
		log.Fatalln("scaning imgs", err.Error())
	}
	for _, entry := range entries {
		pngfile := entry.Name()
		if !strings.HasSuffix(pngfile, ".png") {
			continue
		}
		basename := strings.TrimSuffix(pngfile, ".png")
		yamlFile := internal.Filename(basename, "yaml")
		_ = yamlFile

		imgPath := strings.Join([]string{imgDir, pngfile}, "/")
		metadataPath := strings.Join([]string{imgDir, yamlFile}, "/")

		if _, err := os.Stat(metadataPath); err != nil {

		}

		data, err := os.ReadFile(imgPath)
		if panicker.HasError(err) {
			continue
		}

		oStat.addImg(basename, data)
	}

	fmt.Print(oStat.stats.inform())

	for y := range g4x4.y {
		for x := range g4x4.x {
			spotName := spotName(x, y)
			index := spotIdx(x, y)
			if index >= len(oStat.imageIds) {
				oStat.spotHolder[spotName] = emptySpot
				continue
			}

			var imgId = oStat.imageIds[index]

			oStat.imageSpots[imgId] = spotName
			oStat.spotHolder[spotName] = imgId
		}
	}

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
func (oth *OtherState) addImg(id string, imgData []byte) {
	oth.stats.imgMemSize += int64(len(imgData))
	oth.stats.imgNum += 1

	oth.imageIds = append(oth.imageIds, id)
	oth.imageData[id] = ImgData{
		meta:  emptyMetadata,
		bytes: imgData,
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
