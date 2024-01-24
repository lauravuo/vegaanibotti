package img

import (
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
	"os"
	"strings"

	"github.com/golang/freetype/truetype"
	"github.com/lainio/err2/try"
	"github.com/lauravuo/vegaanibotti/blog/base"
	resize "golang.org/x/image/draw"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

const (
	dpi         = 72.0 // screen resolution in Dots Per Inch
	fontFile    = "./font/Amatic_SC/AmaticSC-Bold.ttf"
	regFontFile = "./font/Amatic_SC/AmaticSC-Regular.ttf"
	spacing     = 0.9 // line spacing (e.g. 2 means double spaced)
)

func getParts(str string, maxW int, drawer *font.Drawer) []string {
	const maxRows = 4

	parts := strings.Split(str, " ")
	res := make([]string, 0)

	for _, part := range parts {
		partLen := drawer.MeasureString(part)
		if partLen > fixed.I(maxW) {
			extraParts := strings.Split(part, "-")
			if len(extraParts) == 1 {
				extraParts = strings.Split(part, "\u00ad")
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

func writeWithFont(img draw.Image, text string) {
	const (
		fontSizeFactor = 6
		middleFactor   = 2
		margin         = 10 * 2
		dpiFactor      = dpi / 72
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

	parts := getParts(text, img.Bounds().Dx()-margin, drawer)
	diffY := int(math.Ceil(size * spacing * dpiFactor))
	coordY := 10 + int(math.Ceil(size*dpiFactor))
	coordY += int((float64(img.Bounds().Dy()) - (size * float64(len(parts)))) / middleFactor)

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

func GenerateThumbnail(post *base.Post, src, target string) {
	const smallFactor = 3

	img := getImageFromFilePath(src)

	if drawImg, ok := img.(draw.Image); ok {
		myGray := color.RGBA{0, 0, 0, 100} //  R, G, B, Alpha
		draw.Draw(drawImg, drawImg.Bounds(), &image.Uniform{myGray}, image.Point{}, draw.Over)
		writeWithFont(drawImg, post.Title)

		out := try.To1(os.Create(target + ".png"))
		defer out.Close()

		try.To(png.Encode(out, drawImg))

		rect := image.Rect(0, 0, drawImg.Bounds().Max.X/smallFactor, drawImg.Bounds().Max.Y/smallFactor)
		dst := image.NewRGBA(rect)
		resize.CatmullRom.Scale(dst, rect, drawImg, drawImg.Bounds(), draw.Over, nil)

		outThumb := try.To1(os.Create(target + "_small.png"))
		defer outThumb.Close()

		try.To(png.Encode(outThumb, dst))
	}
}
