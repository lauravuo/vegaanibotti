package img

import (
	"context"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/golang/freetype/truetype"
	"github.com/lainio/err2/try"
	"github.com/lauravuo/vegaanibotti/blog/base"
	resize "golang.org/x/image/draw"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

const (
	dpi          = 72.0 // screen resolution in Dots Per Inch
	boldFontFile = "./font/Amatic_SC/AmaticSC-Bold.ttf"
	regFontFile  = "./font/Amatic_SC/AmaticSC-Regular.ttf"
	spacing      = 0.9 // line spacing (e.g. 2 means double spaced)
)

//nolint:gocognit,gocyclo,cyclop
func getParts(str, delimiter string, maxW int, drawer *font.Drawer) []string {
	const maxRows = 4

	if delimiter == "" {
		return []string{str}
	}

	parts := strings.Split(str, delimiter)
	res := make([]string, 0)

	for _, part := range parts {
		partLen := drawer.MeasureString(part)

		//nolint:nestif
		if partLen > fixed.I(maxW) {
			extraParts := strings.Split(part, "-")
			if strings.Contains(part, "-") {
				for i := range extraParts {
					if i < len(extraParts)-1 {
						extraParts[i] += "-"
					}
				}
			}

			if len(extraParts) == 1 && strings.Contains(part, "\u00ad") {
				extraParts = strings.Split(part, "\u00ad")

				for i := range extraParts {
					if i < len(extraParts)-1 {
						extraParts[i] += "\u00ad"
					}
				}
			}

			res = append(res, extraParts...)
		} else {
			res = append(res, part)
		}
	}

	removeIndex := -1

	if len(res) > maxRows {
		for index, part := range res {
			if part == "–" {
				removeIndex = index
			}
		}
	}

	if removeIndex > 0 {
		orig := res
		res = orig[:removeIndex]
		res = append(res, orig[removeIndex+1:]...)
	}

	itemCount := len(res)
	if itemCount > maxRows {
		itemCount = maxRows
	}

	return res[:itemCount]
}

func writeWithFont(img draw.Image, text, delimiter, fontFile string, fontSizeFactor int, bottomY bool) {
	const (
		middleFactor = 2
		margin       = 10 * 2
		dpiFactor    = dpi / 72
	)

	// Read the font data.
	fontBytes := try.To1(os.ReadFile(fontFile))
	drawFont := try.To1(truetype.Parse(fontBytes))

	size := float64(img.Bounds().Dy() / fontSizeFactor)

	// Draw the text.
	hinting := font.HintingNone
	drawer := &font.Drawer{
		Dst: img,
		Src: image.NewUniform(image.White),
		Face: truetype.NewFace(drawFont, &truetype.Options{
			Size:    size,
			DPI:     dpi,
			Hinting: hinting,
		}),
	}

	parts := getParts(text, delimiter, img.Bounds().Dx()-margin, drawer)
	diffY := int(math.Ceil(size * spacing * dpiFactor))
	coordY := 10 + int(math.Ceil(size*dpiFactor))

	if !bottomY {
		// center text
		coordY += int((float64(img.Bounds().Dy()) - (size * float64(len(parts)))) / middleFactor)
	} else {
		// text to bottom
		coordY += int((float64(img.Bounds().Dy()) - size - margin*2))
	}

	for _, part := range parts {
		drawer.Dot = fixed.Point26_6{
			X: (fixed.I(img.Bounds().Dx()) - drawer.MeasureString(part)) / 2,
			Y: fixed.I(coordY),
		}
		drawer.DrawString(part)

		coordY += diffY
	}
}

func getImageFromFilePath(filePath string) image.Image {
	f := try.To1(os.Open(filePath))
	defer f.Close()
	img, _ := try.To2(image.Decode(f))

	return img
}

func GenerateThumbnail(post *base.Post, src, target string) (imagePath, smallImagePath string) {
	const smallFactor = 3

	img := getImageFromFilePath(src)

	if drawImg, ok := img.(draw.Image); ok {
		myGray := color.RGBA{0, 0, 0, 100} //  R, G, B, Alpha
		draw.Draw(drawImg, drawImg.Bounds(), &image.Uniform{myGray}, image.Point{}, draw.Over)

		fontSizeFactor := 6
		writeWithFont(drawImg, post.Title, " ", boldFontFile, fontSizeFactor, false)

		fontSizeFactor = 22
		writeWithFont(drawImg, "="+post.Author+"=", "", regFontFile, fontSizeFactor, true)

		imagePath = target + ".png"
		smallImagePath = target + "_small.png"

		out := try.To1(os.Create(imagePath))
		defer out.Close()

		try.To(png.Encode(out, drawImg))

		rect := image.Rect(0, 0, drawImg.Bounds().Max.X/smallFactor, drawImg.Bounds().Max.Y/smallFactor)
		dst := image.NewRGBA(rect)
		resize.CatmullRom.Scale(dst, rect, drawImg, drawImg.Bounds(), draw.Over, nil)

		outThumb := try.To1(os.Create(smallImagePath))
		defer outThumb.Close()

		try.To(png.Encode(outThumb, dst))
	}

	return imagePath, smallImagePath
}

func UploadToCloud(filePaths []string) []string {
	bucketName := os.Getenv("CLOUD_BUCKET_NAME")
	accountID := os.Getenv("CLOUD_ACCOUNT_ID")
	accessKeyID := os.Getenv("CLOUD_ACCESS_KEY_ID")
	accessKeySecret := os.Getenv("CLOUD_ACCESS_KEY_SECRET")
	r2Resolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL: fmt.Sprintf("https://%s.r2.cloudflarestorage.com", accountID),
		}, nil
	})

	cfg := try.To1(config.LoadDefaultConfig(context.TODO(),
		config.WithEndpointResolverWithOptions(r2Resolver),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKeyID, accessKeySecret, "")),
		config.WithRegion("auto"),
	))

	client := s3.NewFromConfig(cfg)
	now := time.Now()
	year := fmt.Sprintf("%d", now.Year())
	month := fmt.Sprintf("%02d", now.Month())
	date := fmt.Sprintf("%s-%s-%02d", year, month, now.Day())

	res := make([]string, 0)

	for _, filePath := range filePaths {
		file := try.To1(os.Open(filePath))
		small := ""

		if strings.Contains(filePath, "small") {
			small = "_small"
		}

		contentType := "image/png"
		path := year + "/" + month + "/" + date + small + ".png"

		_ = try.To1(client.PutObject(context.TODO(), &s3.PutObjectInput{
			Bucket:      aws.String(bucketName),
			Key:         aws.String(path),
			Body:        file,
			ContentType: &contentType,
		}))

		file.Close()

		res = append(res, path)
	}

	return res
}
