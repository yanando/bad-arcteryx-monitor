package monitor

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gtuk/discordwebhook"
)

type ColourOptions struct {
	Selected string         `json:"selected"`
	Options  []ColourOption `json:"options"`
}

type ColourOption struct {
	Label string `json:"label"`
	Value string `json:"value"`
	Image struct {
		Label       string `json:"label"`
		Pathname    string `json:"pathname"`
		URL         string `json:"url"`
		ColourLabel string `json:"colourLabel"`
	} `json:"image"`
	Thumbnail struct {
		Label       string `json:"label"`
		Pathname    string `json:"pathname"`
		URL         string `json:"url"`
		ColourLabel string `json:"colourLabel"`
	} `json:"thumbnail"`
	HexCode       string `json:"hexCode"`
	PrimaryColour string `json:"primaryColour"`
	IsDark        bool   `json:"isDark"`
}

type SizeOptions struct {
	Selected interface{} `json:"selected"`
	Options  []struct {
		Label string `json:"label"`
		Value string `json:"value"`
	} `json:"options"`
}

type Variants []struct {
	ID        string `json:"id"`
	Upc       string `json:"upc"`
	ColourID  string `json:"colourId"`
	SizeID    string `json:"sizeId"`
	Inventory int    `json:"inventory"`
}

type Monitor struct {
	Link    string
	Webhook string

	colourOptions ColourOptions
	variants      Variants
	sizeMap       map[string]string
	initted       bool
}

func (m *Monitor) Monitor() {
	m.sizeMap = make(map[string]string)
	err := m.checkStock()

	if err != nil {
		log.Fatalf("initializing monitor: %s", err)
	}

	m.initted = true

	for {
		time.Sleep(time.Second * 8)
		err = m.checkStock()

		if err != nil {
			log.Printf("error checking stock: %s", err)
		}
	}
}

func (m *Monitor) checkStock() (err error) {
	resp, err := http.Get(m.Link)

	if err != nil {
		return fmt.Errorf("checkstock: %w", err)
	} else if resp.StatusCode != 200 {
		return fmt.Errorf("invalid status code: %d", resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)

	if err != nil {
		return fmt.Errorf("checkstock: %w", err)
	}

	resp.Body.Close()

	if !m.initted {
		colourOptionsString, err := extract(string(bodyBytes), `\"colourOptions\":`, `,\"sizeOptions\"`)

		if err != nil {
			return fmt.Errorf("checkstock: error getting ColourOptions: %w", err)
		}

		err = json.Unmarshal([]byte(strings.ReplaceAll(colourOptionsString, `\"`, `"`)), &m.colourOptions)

		if err != nil {
			return fmt.Errorf("checkstock: error parsing ColourOptions: %w", err)
		}

		sizeOptionsString, err := extract(string(bodyBytes), `\"sizeOptions\":`, `,\"variants\"`)

		if err != nil {
			return fmt.Errorf("checkstock: error getting SizeOptions: %w", err)
		}

		var sizeOptions SizeOptions

		err = json.Unmarshal([]byte(strings.ReplaceAll(sizeOptionsString, `\"`, `"`)), &sizeOptions)

		if err != nil {
			return fmt.Errorf("checkstock: error parsing SizeOptions: %w", err)
		}

		for _, i := range sizeOptions.Options {
			m.sizeMap[i.Value] = i.Label
		}
	}

	variantsString, err := extract(string(bodyBytes), `\"variants\":`, `,\"isReturnable\":`)

	if err != nil {
		return fmt.Errorf("checkstock: error getting variants: %w", err)
	}

	var variants Variants

	err = json.Unmarshal([]byte(strings.ReplaceAll(variantsString, `\"`, `"`)), &variants)

	if err != nil {
		return fmt.Errorf("checkstock: error parsing variants: %w", err)
	}

	m.variants = variants

	// shitty way to do this but oh well

	// use a map of empty structs to avoid duplicate entries
	var changedVars = make(map[string]struct{})
	var changed bool

mainloop:
	for _, i := range variants {
		for _, j := range m.variants {
			// stock change
			if i.ID == j.ID && i.Inventory != j.Inventory && i.Inventory > 0 {
				// add variant to changedVars
				changedVars[i.ColourID] = struct{}{}
				changed = true
			} else if i.ID == j.ID {
				continue mainloop
				// same stock as before
			}
		}
	}

	if !changed {
		log.Println("no change detected")
		return nil
	}

	log.Println("IN STOCK SIZES FOUND")

	log.Println("sending webhook")
	// loop over changed items
	for k := range changedVars {
		var colourItem ColourOption
		for _, i := range m.colourOptions.Options {
			if i.Value == k {
				colourItem = i
			}
		}

		var sizesString string

		for _, i := range m.variants {
			if i.ColourID == colourItem.Value {
				sizesString += fmt.Sprintf("%s [%d]\n", m.sizeMap[i.SizeID], i.Inventory)
			}
		}

		sizesString = strings.TrimSuffix(sizesString, "\n")
		bruh := "Sizes"
		footerText := "made by yanando#0001"

		usr := "ARCY MONITOR"

		msg := discordwebhook.Message{
			Username: &usr,
			Embeds: &[]discordwebhook.Embed{{
				Title:     &colourItem.Image.Label,
				Url:       &m.Link,
				Fields:    &[]discordwebhook.Field{{Name: &bruh, Value: &sizesString}},
				Thumbnail: &discordwebhook.Thumbnail{Url: &colourItem.Image.URL},
				Footer: &discordwebhook.Footer{
					Text: &footerText,
				},
			}},
		}

		discordwebhook.SendMessage(m.Webhook, msg)
	}

	return nil
}

func extract(s, sep1, sep2 string) (string, error) {
	indexSep1 := strings.Index(s, sep1)
	indexSep2 := strings.Index(s, sep2)

	if indexSep1 == -1 {
		return "", errors.New("sep1 doesn't exist in s")
	} else if indexSep2 == -1 {
		return "", errors.New("sep2 doesn't exist in s")
	} else if strings.LastIndex(s, sep2) < indexSep1 {
		return "", errors.New("sep2 doesn't exist after sep1")
	}

	indexSep2 = strings.Index(s[indexSep1+len(sep1):], sep2)

	return s[indexSep1+len(sep1) : indexSep2+indexSep1+len(sep1)], nil
}
