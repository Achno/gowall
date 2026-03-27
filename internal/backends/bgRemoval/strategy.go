package bgremoval

import "image"

type BgRemovalStrategy interface {
	Remove(image.Image) (image.Image, error)
}
