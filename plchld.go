package main

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
	plcHldG      = 0
	plcHldB      = 255
	plcHldA      = 255
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("You have to specify a srcRoot to a directory.")
		return
	}

	srcRoot := os.Args[1]

	// TODO: Add single-file functionality
	if id, err := isDir(srcRoot); err != nil {
		fmt.Println("The provided srcRoot does not exist.")
		return
	} else if !id {
		fmt.Println("The provided srcRoot is not a directory.")
		return
	}

	plcHldRoot := filepath.Join(filepath.Dir(srcRoot), filepath.Base(srcRoot) + plcHldDirAdd)
	fmt.Println("Placeholder:", plcHldRoot)
	if err := os.MkdirAll(plcHldRoot, os.ModePerm); err != nil {
		fmt.Println("Unable to make the placeholder directory.")
		return
	}
	if err := filepath.Walk(srcRoot, func(p string, i os.FileInfo, e error) error {
		return plcHldify(srcRoot, plcHldRoot, p)
	}); err != nil {
		fmt.Println("An error occurred while working on the srcRoot.", err.Error())
		return
	}

	cullStumps(plcHldRoot)

	fmt.Println("Done!")
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
		ext := strings.TrimLeft(strings.ToLower(filepath.Ext(srcPath)), ".")
		switch ext {
		case "jpg", "jpeg", "png":
			iinfo, _, err := image.DecodeConfig(file)
			if err != nil {
				return err
			}
			img := plcHldImg(iinfo.Width, iinfo.Height)
			//fmt.Println(plcHldPath)
			newFile, err := os.OpenFile(plcHldPath, os.O_RDWR | os.O_CREATE, os.ModePerm)
			if err != nil {
				return err
			}
			defer func() {
				if err := newFile.Close(); err != nil {
					fmt.Println("Error writing file \"", plcHldPath, "\":", err.Error())
				}
			}()
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

func cull(path string, info os.FileInfo, err error) error {
	// Removes if empty, otherwise throws an error
	// Calls itself on the removed directory's parent, in case it just removed the last child
	if id, err := isDir(path); err == nil && id {
		if err := os.Remove(path); err == nil {
			cull(filepath.Dir(path), nil, nil)
		}
	} else if err != nil {
		return err
	}
	return nil
}

func cullStumps(root string) {
	filepath.Walk(root, cull)
}