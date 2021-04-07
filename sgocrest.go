package main

import (
	"fmt"
	"image/jpeg"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gabriel-vasile/mimetype"
	"github.com/gen2brain/go-fitz"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/otiai10/gosseract"
	"github.com/thanhpk/randstr"
	"gocv.io/x/gocv"
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
		u.Message = "hello borld!"
		return c.JSON(http.StatusOK, u)
	})
	e.POST("/read/image", ocrImage)
	e.POST("/read/file", ocrFile)
	e.Logger.Fatal(e.Start(":1234"))
}

func ocrImage(c echo.Context) error {
	// User ID from path `users/:id`
	// id := c.Param("id")
	// return c.String(http.StatusOK, "OK\n")

	image, err := c.FormFile("data")
	if err != nil {
		return err
	}
	src, err := image.Open()
	if err != nil {
		return err
	}
	defer src.Close()
	uid := randstr.Hex(16)
	fn := filepath.Join("upload", "image")
	fn = filepath.Join(fn, uid+image.Filename)
	dst, err := os.Create(fn)
	if err != nil {
		return err
	}
	defer dst.Close()
	if _, err = io.Copy(dst, src); err != nil {
		return err
	}
	u := new(r)
	u.Message = ""
	u.Data = ""
	if fileExists(fn) {
		allowed := []string{"image/jpeg", "image/png"}
		mime, _ := mimetype.DetectFile(fn)
		if mimetype.EqualsAny(mime.String(), allowed...) {
			path := filepath.Join(fn)
			im := gocv.IMRead(path, gocv.IMReadColor)
			if im.Empty() {
				u.Message = "Failed (Path)"
			} else {
				gray := gocv.NewMat()
				gocv.CvtColor(im, &gray, gocv.ColorBGRToGray)
				blur := gocv.NewMat()
				gocv.BilateralFilter(gray, &blur, 13, 15, 15)
				scale := gocv.NewMat()
				gocv.ConvertScaleAbs(blur, &scale, 1.5, 25)
				gocv.IMWrite(fn, scale)
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

func ocrFile(c echo.Context) error {
	file, err := c.FormFile("data")
	if err != nil {
		return err
	}
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()
	uid := randstr.Hex(16)
	fn := filepath.Join("upload", "pdf")
	fn = filepath.Join(fn, uid+file.Filename)
	dst, err := os.Create(fn)
	if err != nil {
		return err
	}
	defer dst.Close()
	if _, err = io.Copy(dst, src); err != nil {
		return err
	}
	u := new(r)
	u.Message = ""
	u.Data = ""
	if fileExists(fn) {
		allowed := []string{"application/pdf"}
		mime, _ := mimetype.DetectFile(fn)
		if mimetype.EqualsAny(mime.String(), allowed...) {
			doc, err := fitz.New(fn)
			if err != nil {
				panic(err)
			}
			defer doc.Close()
			for n := 0; n < doc.NumPage(); n++ {
				img, err := doc.Image(n)
				if err != nil {
					panic(err)
				}
				path := filepath.Join("upload", "pdf", "image")
				path = filepath.Join(path, fmt.Sprintf(uid+file.Filename+".%03d.jpg", n))
				f, err := os.Create(path)
				if err != nil {
					panic(err)
				}
				err = jpeg.Encode(f, img, &jpeg.Options{jpeg.DefaultQuality})
				if err != nil {
					panic(err)
				}
				f.Close()
				im := gocv.IMRead(path, gocv.IMReadColor)
				if im.Empty() {
					u.Message = "Failed (Path)"
				} else {
					gray := gocv.NewMat()
					gocv.CvtColor(im, &gray, gocv.ColorBGRToGray)
					blur := gocv.NewMat()
					gocv.BilateralFilter(gray, &blur, 13, 15, 15)
					scale := gocv.NewMat()
					gocv.ConvertScaleAbs(blur, &scale, 1.5, 25)
					gocv.IMWrite(path, scale)
					client := gosseract.NewClient()
					client.SetImage(path)
					text, textErr := client.Text()
					defer client.Close()
					if textErr != nil {
						u.Message = "Failed (Text)"
						println(textErr.Error())
					} else {
						u.Data = text
					}
				}
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
