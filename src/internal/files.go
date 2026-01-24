package internal

import "fmt"

func PngFilename(basename string) string {
	return fmt.Sprintf("%s.png", basename)
}
