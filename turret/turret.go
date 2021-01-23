package turret

import (
	"encoding/json"
	"fmt"
	"github.com/serial"
	"gocv.io/x/gocv"
	"image/color"
	"os"
	"../../radar/radar"
)


func BeginDetection() {
	// set to use a video capture device 0
	deviceID := 0

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

	usb, err := radar.GetUSBDevice("/dev/ttyACM0")

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
	defer window.Close()

	black := color.RGBA{
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
	lowerX := float32((width/2) - THRESH_HOLD)
	upperX := float32((width/2) + THRESH_HOLD)

	lowerY := float32((height/2) - THRESH_HOLD)
	upperY := float32((height/2) + THRESH_HOLD)
	fmt.Printf("Lower X %f, Upper X %f, Lower Y %f, Upper Y %f\n", lowerX, upperX, lowerY, upperY)
	// create serial message payload
	payload := make(map[string]int)

	fmt.Printf("start reading camera device: %v\n", deviceID)
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
			gocv.Rectangle(&img, rect, black, 3)

			// get the middle of the rectangle
			/*
			   we want
			   for mid X
			   to be here   max X is here, so midX = ((maxX - minX) / 2)
					|       |
					v       v
			----------------- <- max Y is here
			|				|
			|				|
			|		.		| <- midY = ((maxY - minY) / 2)
			|				|
			|				|
			-----------------

			*/
			middleX := float32(rect.Max.X) - float32((rect.Max.X - rect.Min.X) / 2)
			middleY := float32(rect.Max.Y) - float32((rect.Max.Y - rect.Min.Y) / 2)
			//fmt.Printf("X max %d, X min %d\n", rect.Max.X, rect.Min.X)

			baseServo := 0
			// calibration for X
			 if lowerX < middleX && middleX < upperX {
				fmt.Printf("X COORDINATE IS GOOD!\n")
				 baseServo = 0
			} else if middleX < lowerX {
				fmt.Printf("X: MOVE THE CAMERA RIGHT\n")
				 baseServo = 5
			} else if middleX > upperX {
				fmt.Printf("X: MOVE THE CAMERA LEFT\n")
				 baseServo = -5
			}

			sideServo := 0
			if lowerY < middleY && middleY < upperY {
				fmt.Printf("Y COORDINATE IS GOOD!\n")
				sideServo = 0

			} else if middleY < lowerY {
				fmt.Printf("Y: MOVE THE CAMERA DOWN\n")
				sideServo = -5

			} else if middleY > upperY {
				fmt.Printf("Y: MOVE THE CAMERA UP\n")
				sideServo = 5
			}

			payload["base"] = baseServo
			payload["side"] = sideServo

			_ = SendData()
		}

		window.IMShow(img)
		window.WaitKey(1)


	}
}

// sends the data over the usb port
func SendData(usb *serial.Port, data *map[string]string) error {

	bytes, err := json.Marshal(*data)
	if err != nil {
		return err
	}

	// append null byte so arduino can terminate
	bytes = append(bytes, '\000')
	_, err = (*usb).Write(bytes)
	if err != nil {
		return err
	}
	return nil

}

func calculateMiddle(end int, start int) float32 {

	return float32((end - start) / 2)

}

func OpenPort(file string) (*os.File, error) {
	return os.Open(file)
}