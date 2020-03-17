package placeholderify

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"
)

const (
	plcHldDirAdd = ".plcHld"
	plcHldR      = 255
	plcHldG      = 255
	plcHldB      = 0
	plcHldA      = 255

)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("You have to specify a srcPath to a directory.")
		return
	}

	srcPath := os.Args[1]

	// TODO: Add single-file functionality
	if id, err := isDir(srcPath); err != nil {
		fmt.Println("The provided srcPath does not exist.")
		return
	} else if !id {
		fmt.Println("The provided srcPath is not a directory.")
		return
	}

	plcHldPath := filepath.Join(filepath.Dir(srcPath), filepath.Base(srcPath) + plcHldDirAdd)
	fmt.Println("Placeholder:", plcHldPath)
	if err := os.MkdirAll(plcHldPath, os.ModePerm); err != nil {
		fmt.Println("Unable to make the placeholder directory.")
		return
	}
	if err := filepath.Walk(srcPath, func(p string, i os.FileInfo, e error) error {
		return plcHldify(srcPath, plcHldPath, p)
	}); err != nil {
		fmt.Println("An error occurred while working on the srcPath.", err.Error())
		return
	}
}

func isDir(path string) (bool, error) {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false, err
	}
	return info.IsDir(), nil
}

func toPlcHldPath(srcRoot, plcHldRoot, srcPath string) (string, error) {
	relPath, err := filepath.Rel(srcRoot, srcPath)
	if err != nil {
		return "", err
	}
	return filepath.Join(plcHldRoot, relPath), nil
}

func plcHldify(srcRoot, plcHldRoot, srcPath string) error {
	plcHldPath, err := toPlcHldPath(srcRoot, plcHldRoot, srcPath)
	if err != nil {
		return err
	}

	if id, err := isDir(srcPath); err == nil && id {
		if err := os.MkdirAll(plcHldPath, os.ModePerm); err != nil {
			return err
		}
	} else if err == nil {
		file, err := os.Open(srcPath)
		if err != nil {
			return err
		}
		ext := strings.ToLower(filepath.Ext(srcPath))
		switch ext {
		case "jpg", "jpeg", "png":
			iinfo, _, err := image.DecodeConfig(file)
			if err != nil {
				return err
			}
			img := plcHldImg(iinfo.Width, iinfo.Height)
			newFile, err := os.OpenFile(plcHldPath, os.O_RDWR | os.O_CREATE, os.ModePerm)
			if err != nil {
				return err
			}
			switch ext {
			case "jpg", "jpeg":
				if err := jpeg.Encode(newFile, img, &jpeg.Options{1}); err != nil {
					return err
				}
			case "png":
				if err := png.Encode(newFile, img); err != nil {
					return err
				}
			}
		}
	} else {
		return err
	}

	return nil
}

func plcHldImg(width, height int) image.Image {
	plcHldColour := color.RGBA{plcHldR, plcHldG, plcHldB, plcHldA}
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(img, img.Bounds(), &image.Uniform{plcHldColour}, image.Pt(0, 0), draw.Src)
	return img
}