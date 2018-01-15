package main

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"os"
	"time"

	"github.com/machinebox/sdk-go/facebox"
	"gocv.io/x/gocv"
)

var (
	fbox          *facebox.Client
	faceAlgorithm = "haarcascade_frontalface_default.xml"
	blue          = color.RGBA{0, 0, 255, 0}
)

func main() {
	fbox = facebox.New("http://localhost:8080")

	// open webcam
	webcam, err := gocv.VideoCaptureDevice(0)
	if err != nil {
		log.Fatalf("error opening video capture device: %v", err)
	}
	defer webcam.Close()

	// open display window
	window := gocv.NewWindow("webcamFacebox")
	defer window.Close()

	// prepare image matrix
	img := gocv.NewMat()
	defer img.Close()

	// load classifier to recognize faces
	classifier := gocv.NewCascadeClassifier()
	defer classifier.Close()
	classifier.Load(faceAlgorithm)

	for {
		if ok := webcam.Read(img); !ok {
			log.Print("cannot read webcam")
			continue
		}
		if img.Empty() {
			continue
		}

		// detect faces
		rects := classifier.DetectMultiScale(img)

		for _, r := range rects {
			// Save each found face into the file
			imgFace := img.Region(r)
			imgName := fmt.Sprintf("%d.jpg", time.Now().UnixNano())
			gocv.IMWrite(imgName, imgFace)
			imgFace.Close()

			f, err := os.Open(imgName)
			if err != nil {
				log.Printf("unable to open saved img: %v", err)
			}

			faces, err := fbox.Check(f)
			if err != nil {
				log.Printf("unable to recognize face: %v", err)
			}

			f.Close()
			// gocv requires us to save file, so we need to remove it here
			os.Remove(imgName)

			var caption = "I DO NOT know you"
			if len(faces) > 0 {
				caption = fmt.Sprintf("I know you %s", faces[0].Name)
			}

			// print caption and draw rectangle for the face
			size := gocv.GetTextSize(caption, gocv.FontHersheyPlain, 3, 2)
			pt := image.Pt(r.Min.X+(r.Min.X/2)-(size.X/2), r.Min.Y-2)
			gocv.PutText(img, caption, pt, gocv.FontHersheyPlain, 3, blue, 2)
			gocv.Rectangle(img, r, blue, 3)
		}

		// show the image in the window, and wait 500ms
		window.IMShow(img)
		window.WaitKey(500)
	}
}
