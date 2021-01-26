package turret

import (
	"../../radar/radar"
	"fmt"
	"github.com/serial"
	"gocv.io/x/gocv"
	"image/color"
	"time"
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

	usb, err := radar.GetUSBDevice("/dev/ttyUSB0")

	if err != nil {
		fmt.Printf("Error in opening USB device %s\v", err)
		return
	}

	//radar.RecvDevice(usb, buffer, '\n')
	// open webcam
	webcam, err := gocv.OpenVideoCapture(deviceID)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer webcam.Close()

	// prepare image matrix
	img := gocv.NewMat()
	defer img.Close()

	classifier := gocv.NewCascadeClassifier()
	if !classifier.Load("data/haarcascade_frontalface_default.xml") {
		fmt.Println("Error reading cascade file: data/haarcascade_frontalface_default.xml")
		return
	}

	// open display window
	window := gocv.NewWindow("Face Detect")
	window.ResizeWindow(640, 480)
	defer window.Close()

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
	// calculate upper and lower bounds of width and height of the webcam frame
	lowerX := float32((width / 2) - THRESH_HOLD)
	upperX := float32((width / 2) + THRESH_HOLD)

	lowerY := float32((height / 2) - THRESH_HOLD)
	upperY := float32((height / 2) + THRESH_HOLD)
	fmt.Printf("Lower X %f, Upper X %f, Lower Y %f, Upper Y %f\n", lowerX, upperX, lowerY, upperY)
	fmt.Printf("start reading camera device: %v\n", deviceID)

	moveBy := 1
	for {
		if ok := webcam.Read(&img); !ok {
			fmt.Printf("cannot read device %v\n", deviceID)

			return
		}
		if img.Empty() {
			continue
		}

		faces := classifier.DetectMultiScale(img)
		for _, rect := range faces {
			gocv.Rectangle(&img, rect, white, 3)

			// get the middle of the rectangle
			/*
							   we want
							   for mid X
							   to be here   max X is here, so midX = ((maxX - minX) / 2)
			                        |       |
			                        v       v
							----------------- <- max Y is here
							|               |
							|               |
							|       .       | <- midY = ((maxY - minY) / 2)
							|               |
							|               |
							-----------------

			*/
			middleX := float32(rect.Max.X) - float32((rect.Max.X-rect.Min.X)/2)
			middleY := float32(rect.Max.Y) - float32((rect.Max.Y-rect.Min.Y)/2)
			fmt.Printf("Mid X %f, Mid Y %f, one foot away\n", middleX, middleY)
			fmt.Printf("Number of pixels across X, one foot away %d, number of pixels across Y, one foot away %d\n", rect.Max.X-rect.Max.Y, rect.Max.Y-rect.Min.Y)

			baseServo := 0
			// calibration for X
			if lowerX < middleX && middleX < upperX {
				//fmt.Printf("X COORDINATE IS GOOD!\n")
				baseServo = 0
			} else if middleX < lowerX {
				//fmt.Printf("X: MOVE THE CAMERA RIGHT\n")
				baseServo = moveBy
			} else if middleX > upperX {
				//fmt.Printf("X: MOVE THE CAMERA LEFT\n")
				baseServo = -moveBy
			}

			sideServo := 0
			if lowerY < middleY && middleY < upperY {
				//fmt.Printf("Y COORDINATE IS GOOD!\n")
				sideServo = 0

			} else if middleY < lowerY {
				//fmt.Printf("Y: MOVE THE CAMERA DOWN\n")
				sideServo = -moveBy

			} else if middleY > upperY {
				//fmt.Printf("Y: MOVE THE CAMERA UP\n")
				sideServo = moveBy
			}

			_ = SendData(usb, baseServo, sideServo)
			break
		}

		window.IMShow(img)
		window.WaitKey(1)
		time.Sleep(1)

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

// calculates the angle needed to turn in the x direction
func calculateRotationX(hitPosition int, imageCenter int, imageWidth int) int {

	// calculate total pixels needed to turn to the hit position
	totalPixels := imageCenter - hitPosition
	//sign := totalPixels < 0

	// translate pixels to an angle to turn, defaulting to 1 if too low
	// assuming 1 meter from center of camera

	return totalPixels

}

// calculates the angle needed to turn in the y direction
