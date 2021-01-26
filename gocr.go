package main

import (
	"io"
	"net/http"
	"os"

	"github.com/gabriel-vasile/mimetype"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/otiai10/gosseract"
	"github.com/thanhpk/randstr"
)

func main() {
	e := echo.New()
	e.Use(middleware.GzipWithConfig(middleware.GzipConfig{
		Level: 5,
	}))
	e.Use(middleware.SecureWithConfig(middleware.SecureConfig{
		HSTSMaxAge:            3600,
		ContentSecurityPolicy: "default-src 'self'",
		XSSProtection:         "1; mode=block",
		ContentTypeNosniff:    "nosniff",
		XFrameOptions:         "SAMEORIGIN",
	}))
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: `{"time":"${time_rfc3339_nano}","remote_ip":"${remote_ip}",` +
			`"method":"${method}","uri":"${uri}","status":${status},"error":"${error}",` +
			`"latency_human":"${latency_human}","bytes_in":${bytes_in},` +
			`"bytes_out":${bytes_out}}` + "\n",
	}))

	e.GET("/", func(c echo.Context) error {
		u := new(s)
		u.Message = "hello bor!"
		return c.JSON(http.StatusOK, u)
	})

	e.POST("/read", ocr)

	e.Logger.Fatal(e.Start(":1234"))
}

func ocr(c echo.Context) error {
	// User ID from path `users/:id`
	// id := c.Param("id")
	// return c.String(http.StatusOK, "OK\n")

	// Get
	image, err := c.FormFile("image")
	if err != nil {
		return err
	}

	// Source
	src, err := image.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	// Destination
	uid := randstr.Hex(16)
	// uuid := uuid.Must(uuid.NewV4())
	// spew.Dump(uuid)

	fn := "upload/" + uid + image.Filename

	dst, err := os.Create(fn)
	if err != nil {
		return err
	}
	defer dst.Close()

	// Copy
	if _, err = io.Copy(dst, src); err != nil {
		return err
	}

	u := new(r)
	u.Message = ""
	u.Data = ""

	if fileExists(fn) {
		allowed := []string{"image/jpeg", "image/png", "application/pdf"}
		// allowed := []string{"image/jpeg", "image/png"}
		mime, _ := mimetype.DetectFile(fn)
		if mimetype.EqualsAny(mime.String(), allowed...) {
			if mime.Extension() == ".pdf" {
				u.Message = "Failed (pdf)"

				// doc, err := fitz.New(fn)
				// if err != nil {
				// 	panic(err)
				// }

				// defer doc.Close()

				// tmpDir, err := ioutil.TempDir(os.TempDir(), "fitz")
				// if err != nil {
				// 	panic(err)
				// }

				// // Extract pages as images
				// for n := 0; n < doc.NumPage(); n++ {
				// 	img, err := doc.Image(n)
				// 	if err != nil {
				// 		panic(err)
				// 	}

				// 	f, err := os.Create(filepath.Join(tmpDir, fmt.Sprintf("test%03d.jpg", n)))
				// 	if err != nil {
				// 		panic(err)
				// 	}

				// 	err = jpeg.Encode(f, img, &jpeg.Options{jpeg.DefaultQuality})
				// 	if err != nil {
				// 		panic(err)
				// 	}

				// 	f.Close()
				// }
			} else {
				//run cv2 from python
				// cmd := exec.Command("bash", "ocv.sh", fn)
				// out, err := cmd.Output()
				// if err != nil {
				// u.Message = "Failed" + err.Error()
				// }

				// path := filepath.Join(fn)
				// im := gocv.IMRead(path, gocv.IMReadColor)
				// if im.Empty() {
				// u.Message = "Failed (Path)"
				// } else {
				// gray := gocv.NewMat()
				// gocv.CvtColor(im, &gray, gocv.ColorBGRToGray)
				// blur := gocv.NewMat()
				// gocv.BilateralFilter(gray, &blur, 13, 15, 15)
				// scale := gocv.NewMat()
				// gocv.ConvertScaleAbs(blur, &scale, 1.5, 25)

				// gocv.IMWrite(fn, scale)

				client := gosseract.NewClient()
				client.SetImage(fn)
				text, textErr := client.Text()
				defer client.Close()

				if textErr != nil {
					u.Message = "Failed (Text)"
					println(textErr.Error())
				} else {
					u.Data = text
				}
				// }
			}
		} else {
			os.Remove(fn)
			u.Message = "Failed (Mime)"
		}
	} else {
		u.Message = "Failed (File)"
	}

	return c.JSON(http.StatusOK, u)
}

type r struct {
	Message string `json:"message" xml:"message" form:"message" query:"message"`
	Data    string `json:"data" xml:"data" form:"data" query:"data"`
}

type s struct {
	Message string `json:"message" xml:"message" form:"message" query:"message"`
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
