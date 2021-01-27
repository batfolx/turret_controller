package turret

import (
	"fmt"
	"github.com/batfolx/radar"
	"github.com/tarm/serial"
	"gocv.io/x/gocv"
	"image"
	"image/color"
	"math"
)

func BeginDetection() {
	// set to use a video capture device 0
	deviceID := 1

	// open webcam
	webcam, err := gocv.OpenVideoCapture(deviceID)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer webcam.Close()

	// open display window
	window := gocv.NewWindow("Face Detect")
	defer window.Close()

	// prepare image matrix
	img := gocv.NewMat()
	defer img.Close()

	// color for the rect when faces detected
	blue := color.RGBA{0, 0, 255, 0}

	// load classifier to recognize faces
	classifier := gocv.NewCascadeClassifier()
	defer classifier.Close()

	if !classifier.Load("data/haarcascade_frontalface_default.xml") {
		fmt.Println("Error reading cascade file: data/haarcascade_frontalface_default.xml")
		return
	}

	fmt.Printf("start reading camera device: %v\n", deviceID)
	for {
		if ok := webcam.Read(&img); !ok {
			fmt.Printf("cannot read device %v\n", deviceID)
			return
		}
		if img.Empty() {
			continue
		}

		// detect faces
		rects := classifier.DetectMultiScale(img)
		fmt.Printf("found %d faces\n", len(rects))

		// draw a rectangle around each face on the original image
		for _, r := range rects {
			gocv.Rectangle(&img, r, blue, 3)
		}

		// show the image in the window, and wait 1 millisecond
		window.IMShow(img)
		window.WaitKey(1)
	}
}

func BeginDetectionHeadless() {
	// set to use a video capture device 0
	deviceID := 0
	device := "/dev/ttyUSB0"

	fmt.Printf("Opening device %s...\n", device)
	usb, err := radar.GetUSBDevice(device)

	if err != nil {
		fmt.Printf("Error in opening USB device, %s %s\v", device, err)
		return
	}

	fmt.Printf("USB successfully opened!\nOpening webcam...\n")
	// open webcam
	webcam, err := gocv.OpenVideoCapture(deviceID)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer webcam.Close()
	fmt.Printf("Successfully opened webcam! Continuing...\n")

	// prepare image matrix
	img := gocv.NewMat()
	defer img.Close()

	classifier := gocv.NewCascadeClassifier()
	if !classifier.Load("data/haarcascade_frontalface_default.xml") {
		fmt.Println("Error reading cascade file: data/haarcascade_frontalface_default.xml")
		return
	}

	// open display window
	//window := gocv.NewWindow("Face Detect")
	//window.ResizeWindow(640, 480)
	//defer window.Close()

	white := color.RGBA{
		R: 255,
		G: 255,
		B: 255,
		A: 0,
	}

	_ = webcam.Read(&img)

	if img.Empty() {
		return
	}

	height := webcam.Get(gocv.VideoCaptureFrameHeight)
	width := webcam.Get(gocv.VideoCaptureFrameWidth)
	fmt.Printf("Frame height %f and width %f\n", height, width)

	//webcam.Set(gocv.VideoCaptureFrameWidth, float64(height))
	//webcam.Set(gocv.VideoCaptureFrameHeight, float64(width))

	// calculate upper and lower bounds of width and height of the webcam frame
	lowerX := float32((width / 2) - THRESH_HOLD)
	upperX := float32((width / 2) + THRESH_HOLD)

	lowerY := float32((height / 2) - THRESH_HOLD)
	upperY := float32((height / 2) + THRESH_HOLD)
	fmt.Printf("Lower X %f, Upper X %f, Lower Y %f, Upper Y %f\n", lowerX, upperX, lowerY, upperY)
	fmt.Printf("start reading camera device: %v\n", deviceID)

	minSize := image.Point{}
	maxSize := image.Point{}

	for {
		if ok := webcam.Read(&img); !ok {
			fmt.Printf("cannot read device %v\n", deviceID)
			return
		}
		if img.Empty() {
			continue
		}

		faces := classifier.DetectMultiScaleWithParams(img, 1.3, 3, 0, minSize, maxSize)
		for _, rect := range faces {
			gocv.Rectangle(&img, rect, white, 3)

			// get the middle of the rectangle
			/*
				   we want
				   for mid X
				   to be here   max X is here, so midX = maxX -  ((maxX - minX) / 2)
				        |       |
				        v       v
				----------------- <- max Y is here
				|               |
				|               |
				|       .       | <- midY = maxY - ((maxY - minY) / 2)
				|               |
				|               |
				-----------------

			*/
			middleX := float32(rect.Max.X) - float32((rect.Max.X-rect.Min.X)/2)
			middleY := float32(rect.Max.Y) - float32((rect.Max.Y-rect.Min.Y)/2)
			//fmt.Printf("Mid X %f, Mid Y %f, one foot away\n", middleX, middleY)
			//fmt.Printf("Number of pixels across X, one foot away %d, number of pixels across Y, one foot away %d\n", rect.Max.X-rect.Max.Y, rect.Max.Y-rect.Min.Y)
			degreesX := calculateRotation(int(middleX), int(width/2))
			degreesY := -calculateRotation(int(middleY), int(height/2))

			_ = SendData(usb, degreesX, degreesY)
			break
		}

		//window.IMShow(img)
		//window.WaitKey(1)
		//time.Sleep(1)

	}
}

// sends the data over the usb port
func SendData(usb *serial.Port, base int, side int) error {

	// create bytes payload and send
	bytes := []byte(fmt.Sprintf("{\"base\": %d, \"side\": %d}\000", base, side))
	_, err := (*usb).Write(bytes)
	if err != nil {
		return err
	}
	return nil

}

// calculates the angle needed to turn in the direction
func calculateRotation(hitPosition int, imageCenter int) int {

	// calculate total pixels needed to turn to the hit position
	totalPixels := imageCenter - hitPosition

	// translate pixels to an angle to turn, defaulting to 1 if too low
	// assuming 1 meter from center of camera
	distanceCm := float32(totalPixels) * CM_PER_PIXEL

	// assume 1m distance for now, maybe get this from a sensor on the pi or something
	degrees := (math.Atan(float64(distanceCm)/100) * 180) / math.Pi
	//fmt.Printf("This is degrees %f and distance cm %f\n", degrees, distanceCm)
	value := degrees

	if -MIN_ANGLE < value && value < MIN_ANGLE {
		return 0
	} else {
		return int(value)
	}

}
