package main

import (
	"fmt"
	"github.com/disintegration/imaging"
	"image"
	"image/png"
	_ "image/png"
	"log"
	"os"

	"github.com/oliamb/cutter"
)

func getImageFromFilePath(filePath string) (image.Image, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	i, _, err := image.Decode(f)
	return i, err
}

func getFaceImages(i image.Image) ([]image.Image, error) {
	var faces []image.Image
	for y := 0; y < 3; y++ {
		for x := 0; x < 9; x++ {
			croppedImage, err := cutter.Crop(i, cutter.Config{
				Width:   8,
				Height:  8,
				//Options: cutter.,
				Anchor:  image.Point{X: x * 8, Y: y * 8},
			})
			if err != nil {
				return nil, err
			}
			faces = append(faces, croppedImage)
		}
	}
	return faces, nil
}

func placeFacesOnSkin(faces []image.Image) ([]image.Image, error) {
	var skins []image.Image
	var baseSkin image.Image
	baseSkin, err := getImageFromFilePath("baseSkin.png")
	if err != nil {
		log.Fatal(err)
	}
	for _, face := range faces {
		baseImage := baseSkin
		toAppend := imaging.Paste(baseImage, face, image.Point{X: 8, Y: 8})
		skins = append(skins, toAppend)
	}
	return skins, nil
}

func saveSkins(skins []image.Image) error {
	_, err := os.Stat("output")
	if os.IsNotExist(err) {
		_ = os.Mkdir("output", 0755)
	}
	for i, skin := range skins {
		file, _ := os.Create(fmt.Sprintf("./output/%v.png", i+1))
		_ = png.Encode(file, skin)
		_ = file.Close()
	}
	return nil
}

func main() {
	originalImage, err := getImageFromFilePath("image.png")
	if err != nil {
		log.Fatal(err)
	}
	faces, _ := getFaceImages(originalImage)
	skins, _ := placeFacesOnSkin(faces)
	err = saveSkins(skins)
	if err != nil {
		log.Fatal(err)
	}
}
