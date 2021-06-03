package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/png"
	_ "image/png"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/disintegration/imaging"
	"github.com/oliamb/cutter"
)

type profileResponse struct {
	Capes []interface{} `json:"capes"`
	ID    string        `json:"id"`
	Name  string        `json:"name"`
	Skins []struct {
		ID      string `json:"id"`
		State   string `json:"state"`
		URL     string `json:"url"`
		Variant string `json:"variant"`
	} `json:"skins"`
}

func getImageFromFilePath(filePath string) (image.Image, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	i, _, err := image.Decode(f)
	return i, err
}

func getUuidFromBearer(bearer string) (string, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://api.minecraftservices.com/minecraft/profile", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", bearer))
	res, err := client.Do(req)
	if err != nil {
		return "", err
	}

	var body profileResponse
	json.NewDecoder(res.Body).Decode(&body)
	fmt.Println(body.ID)
	return body.ID, nil
}

func getFaceImages(i image.Image) ([]image.Image, error) {
	var faces []image.Image
	for y := 0; y < 3; y++ {
		for x := 0; x < 9; x++ {
			croppedImage, err := cutter.Crop(i, cutter.Config{
				Width:  8,
				Height: 8,
				Anchor: image.Point{X: x * 8, Y: y * 8},
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
	baseSkin = imaging.Paste(baseSkin, image.NewRGBA(image.Rectangle{image.Point{0, 0}, image.Point{8, 8}}), image.Point{40, 8})
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

func input(msg string) (string, error) {
	fmt.Printf("%v", msg)
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		return scanner.Text(), nil
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}
	return "", errors.New("who knows")
}

func handleErr(err error) {
	fmt.Printf("We ran into an error: %v\n", err)
	input("Press enter to exit: ")
	os.Exit(0)
}

func main() {
	title := `
███╗   ██╗ █████╗ ███╗   ███╗███████╗███╗   ███╗ ██████╗    ███████╗██╗  ██╗██╗███╗   ██╗     █████╗ ██████╗ ████████╗
████╗  ██║██╔══██╗████╗ ████║██╔════╝████╗ ████║██╔════╝    ██╔════╝██║ ██╔╝██║████╗  ██║    ██╔══██╗██╔══██╗╚══██╔══╝
██╔██╗ ██║███████║██╔████╔██║█████╗  ██╔████╔██║██║         ███████╗█████╔╝ ██║██╔██╗ ██║    ███████║██████╔╝   ██║   
██║╚██╗██║██╔══██║██║╚██╔╝██║██╔══╝  ██║╚██╔╝██║██║         ╚════██║██╔═██╗ ██║██║╚██╗██║    ██╔══██║██╔══██╗   ██║   
██║ ╚████║██║  ██║██║ ╚═╝ ██║███████╗██║ ╚═╝ ██║╚██████╗    ███████║██║  ██╗██║██║ ╚████║    ██║  ██║██║  ██║   ██║   
╚═╝  ╚═══╝╚═╝  ╚═╝╚═╝     ╚═╝╚══════╝╚═╝     ╚═╝ ╚═════╝    ╚══════╝╚═╝  ╚═╝╚═╝╚═╝  ╚═══╝    ╚═╝  ╚═╝╚═╝  ╚═╝   ╚═╝   

`
	fmt.Print(title)
	fmt.Println("Generating images...")
	originalImage, err := getImageFromFilePath("image.png")
	if err != nil {
		handleErr(err)
	}
	faces, err := getFaceImages(originalImage)
	if err != nil {
		handleErr(err)
	}
	skins, err := placeFacesOnSkin(faces)
	if err != nil {
		handleErr(err)
	}
	err = saveSkins(skins)
	if err != nil {
		handleErr(err)
	}
	fmt.Println("Generated :D")
	toContinue, _ := input("(y/n) Would you like to apply the skins to your account? ")
	switch strings.ToLower(toContinue) {
	case "y", "yes":
		fmt.Println("before applying these skins. make sure you have NO skins on your namemc profile. you can remove them at \"edit profile\" -> \"skins\"\nMake sure you save any important skins")
		fmt.Println("to apply these skins this program needs your bearer token.\nGuide on getting it: https://kqzz.github.io/mc-bearer-token")
		bearer, _ := input("Bearer? ")
		err = applySkins(bearer, skins)
		if err != nil {
			handleErr(err)
		}
		fmt.Println("DONE! All the skins have been applied to your account.")
	}
}
