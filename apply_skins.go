package main

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"image"
	"image/png"
	"log"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	"time"
)

func newImageUploadRequest(uri string, params map[string]string, paramName string, imFile image.Image) (*http.Request, error) {

	buf := new(bytes.Buffer)
	err := png.Encode(buf, imFile)
	imBytes := buf.Bytes()

	if err != nil {
		log.Fatal(err)
	}

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile(paramName, "skin.png")
	if err != nil {
		return nil, err
	}
	part.Write(imBytes)

	for key, val := range params {
		_ = writer.WriteField(key, val)
	}
	err = writer.Close()
	if err != nil {
		return nil, err
	}

	request, _ := http.NewRequest("POST", uri, body)
	request.Header.Add("Content-Type", writer.FormDataContentType())
	return request, nil
}

func firefoxClient() *http.Client {
	cipherSuites := []uint16{
		0x1301, 0x1303, 0x1302, 0xc02b, 0xc02f, 0xcca9, 0xcca8, 0xc02c, 0xc030, 0xc00a, 0xc009, 0xc013, 0xc014, 0x009c, 0x009d, 0x002f, 0x0035, 0x000a,
	}

	cookieJar, err := cookiejar.New(nil)
	if err != nil {
		panic(err)
	}

	tlsConfig := &tls.Config{CipherSuites: cipherSuites, MaxVersion: tls.VersionTLS12}
	client := &http.Client{Transport: &http.Transport{TLSClientConfig: tlsConfig}, Jar: cookieJar}

	return client
}


func cacheSkin(uuid string) {
	client := firefoxClient()
	_, err := client.Get(fmt.Sprintf("https://namemc.com/profile/%v", uuid))
	if err != nil {
		log.Fatal(err)
	}
}

func applySkins(bearer string, skins []image.Image) error {
	client := &http.Client{}
	for i, j := 0, len(skins)-1; i < j; i, j = i+1, j-1 {
		skins[i], skins[j] = skins[j], skins[i]
	}
	uuid, err := getUuidFromBearer(bearer)
	if err != nil {
		handleErr(err)
	}
	for i, skin := range skins {
		params := map[string]string{
			"variant": "slim",
		}
		request, err := newImageUploadRequest("https://api.minecraftservices.com/minecraft/profile/skins", params, "file", skin)
		if err != nil {
			log.Fatal(err)
		}
		request.Header.Set("Authorization", fmt.Sprintf("Bearer %v", bearer))
		request.Header.Set("file", "skin.png;type=image/png")
		resp, err := client.Do(request)
		if err != nil {
			handleErr(err)
		}

		fmt.Printf("skin #%v | %v\n", i, resp.StatusCode)
		time.Sleep(time.Second * 60)
		cacheSkin(uuid)
		time.Sleep(time.Second * 45)
	}
	return nil
}
