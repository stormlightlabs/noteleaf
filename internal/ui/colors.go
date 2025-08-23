// This module contains colors from github.com/charmbracelet/x/exp/charmtone
//
// See: https://github.com/charmbracelet/x/blob/main/exp/charmtone/charmtone.go
package ui

import (
	"fmt"
	"image/color"
	"slices"

	"github.com/lucasb-eyer/go-colorful"
)

var _ color.Color = Key(0)

// Key is a type for color keys.
type Key int

const (
	Cumin Key = iota
	Tang
	Yam
	Paprika
	Bengal
	Uni
	Sriracha
	Coral
	Salmon
	Chili
	Cherry
	Tuna
	Macaron
	Pony
	Cheeky
	Flamingo
	Dolly
	Blush
	Urchin
	Mochi
	Lilac
	Prince
	Violet
	Mauve
	Grape
	Plum
	Orchid
	Jelly
	Charple
	Hazy
	Ox
	Sapphire
	Guppy
	Oceania
	Thunder
	Anchovy
	Damson
	Malibu
	Sardine
	Zinc
	Turtle
	Lichen
	Guac
	Julep
	Bok
	Mustard
	Citron
	Zest
	Pepper
	BBQ
	Charcoal
	Iron
	Oyster
	Squid
	Smoke
	Ash
	Salt
	Butter

	// Diffs: additions. The brightest color in this set is Julep, defined above.
	Pickle
	Gator
	Spinach

	// Diffs: deletions. The brightest color in this set is Cherry, defined above.
	Pom
	Steak
	Toast

	// Provisional.
	NeueGuac
	NeueZinc
)

func (k Key) Hex() string {
	switch k {
	case Cumin:
		return "#BF976F"
	case Tang:
		return "#FF985A"
	case Yam:
		return "#FFB587"
	case Paprika:
		return "#D36C64"
	case Bengal:
		return "#FF6E63"
	case Uni:
		return "#FF937D"
	case Sriracha:
		return "#EB4268"
	case Coral:
		return "#FF577D"
	case Salmon:
		return "#FF7F90"
	case Chili:
		return "#E23080"
	case Cherry:
		return "#FF388B"
	case Tuna:
		return "#FF6DAA"
	case Macaron:
		return "#E940B0"
	case Pony:
		return "#FF4FBF"
	case Cheeky:
		return "#FF79D0"
	case Flamingo:
		return "#F947E3"
	case Dolly:
		return "#FF60FF"
	case Blush:
		return "#FF84FF"
	case Urchin:
		return "#C337E0"
	case Mochi:
		return "#EB5DFF"
	case Lilac:
		return "#F379FF"
	case Prince:
		return "#9C35E1"
	case Violet:
		return "#C259FF"
	case Mauve:
		return "#D46EFF"
	case Grape:
		return "#7134DD"
	case Plum:
		return "#9953FF"
	case Orchid:
		return "#AD6EFF"
	case Jelly:
		return "#4A30D9"
	case Charple:
		return "#6B50FF"
	case Hazy:
		return "#8B75FF"
	case Ox:
		return "#3331B2"
	case Sapphire:
		return "#4949FF"
	case Guppy:
		return "#7272FF"
	case Oceania:
		return "#2B55B3"
	case Thunder:
		return "#4776FF"
	case Anchovy:
		return "#719AFC"
	case Damson:
		return "#007AB8"
	case Malibu:
		return "#00A4FF"
	case Sardine:
		return "#4FBEFE"
	case Zinc:
		return "#10B1AE"
	case Turtle:
		return "#0ADCD9"
	case Lichen:
		return "#5CDFEA"
	case Guac:
		return "#12C78F"
	case Julep:
		return "#00FFB2"
	case Bok:
		return "#68FFD6"
	case Mustard:
		return "#F5EF34"
	case Citron:
		return "#E8FF27"
	case Zest:
		return "#E8FE96"
	case Pepper:
		return "#201F26"
	case BBQ:
		return "#2d2c35"
	case Charcoal:
		return "#3A3943"
	case Iron:
		return "#4D4C57"
	case Oyster:
		return "#605F6B"
	case Squid:
		return "#858392"
	case Smoke:
		return "#BFBCC8"
	case Ash:
		return "#DFDBDD"
	case Salt:
		return "#F1EFEF"
	case Butter:
		return "#FFFAF1"
	// Diffs: additions.
	case Pickle:
		return "#00A475"
	case Gator:
		return "#18463D"
	case Spinach:
		return "#1C3634"
	// Diffs: deletions.
	case Pom:
		return "#AB2454"
	case Steak:
		return "#582238"
	case Toast:
		return "#412130"
	// Provisional.
	case NeueGuac:
		return "#00b875"
	case NeueZinc:
		return "#0e9996"
	default:
		return ""
	}
}

func (k Key) String() string {
	switch k {
	case Cumin:
		return "Cumin"
	case Tang:
		return "Tang"
	case Yam:
		return "Yam"
	case Paprika:
		return "Paprika"
	case Bengal:
		return "Bengal"
	case Uni:
		return "Uni"
	case Sriracha:
		return "Sriracha"
	case Coral:
		return "Coral"
	case Salmon:
		return "Salmon"
	case Chili:
		return "Chili"
	case Cherry:
		return "Cherry"
	case Tuna:
		return "Tuna"
	case Macaron:
		return "Macaron"
	case Pony:
		return "Pony"
	case Cheeky:
		return "Cheeky"
	case Flamingo:
		return "Flamingo"
	case Dolly:
		return "Dolly"
	case Blush:
		return "Blush"
	case Urchin:
		return "Urchin"
	case Mochi:
		return "Mochi"
	case Lilac:
		return "Lilac"
	case Prince:
		return "Prince"
	case Violet:
		return "Violet"
	case Mauve:
		return "Mauve"
	case Grape:
		return "Grape"
	case Plum:
		return "Plum"
	case Orchid:
		return "Orchid"
	case Jelly:
		return "Jelly"
	case Charple:
		return "Charple"
	case Hazy:
		return "Hazy"
	case Ox:
		return "Ox"
	case Sapphire:
		return "Sapphire"
	case Guppy:
		return "Guppy"
	case Oceania:
		return "Oceania"
	case Thunder:
		return "Thunder"
	case Anchovy:
		return "Anchovy"
	case Damson:
		return "Damson"
	case Malibu:
		return "Malibu"
	case Sardine:
		return "Sardine"
	case Zinc:
		return "Zinc"
	case Turtle:
		return "Turtle"
	case Lichen:
		return "Lichen"
	case Guac:
		return "Guac"
	case Julep:
		return "Julep"
	case Bok:
		return "Bok"
	case Mustard:
		return "Mustard"
	case Citron:
		return "Citron"
	case Zest:
		return "Zest"
	case Pepper:
		return "Pepper"
	case BBQ:
		return "BBQ"
	case Charcoal:
		return "Charcoal"
	case Iron:
		return "Iron"
	case Oyster:
		return "Oyster"
	case Squid:
		return "Squid"
	case Smoke:
		return "Smoke"
	case Ash:
		return "Ash"
	case Salt:
		return "Salt"
	case Butter:
		return "Butter"
	case Pickle:
		return "Pickle"
	case Gator:
		return "Gator"
	case Spinach:
		return "Spinach"
	case Pom:
		return "Pom"
	case Steak:
		return "Steak"
	case Toast:
		return "Toast"
	case NeueGuac:
		return "NeueGuac"
	case NeueZinc:
		return "NeueZinc"
	default:
		return fmt.Sprintf("Key(%d)", int(k))
	}
}

// RGBA returns the red, green, blue, and alpha values of the color for interface [color.Color]
func (k Key) RGBA() (r, g, b, a uint32) {
	c, err := colorful.Hex(k.Hex())
	if err != nil {
		panic(fmt.Sprintf("invalid color key %d: %s: %v", k, k.String(), err))
	}
	return c.RGBA()
}

func (k Key) IsPrimary() bool {
	return slices.Contains(PrimaryColors, k)
}

func (k Key) IsSecondary() bool {
	return slices.Contains(SecondaryColors, k)
}

func (k Key) IsTertiary() bool {
	return slices.Contains(TertiaryColors, k)
}
