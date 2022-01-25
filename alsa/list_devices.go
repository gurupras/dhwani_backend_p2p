package alsa

import (
	"fmt"
	"strings"

	"github.com/gurupras/dhwani_backend_p2p/types"
	log "github.com/sirupsen/logrus"
	"github.com/yobert/alsa"
)

func ListDevicesWithLib() ([]*types.AudioDevice, error) {
	cards, err := alsa.OpenCards()
	if err != nil {
		return nil, err
	}
	log.Debugf("Got %v cards\n", len(cards))
	results := make([]*types.AudioDevice, 0)
	for _, card := range cards {
		cardDevices, err := card.Devices()
		cardInfo := make([]string, 0)
		if err != nil {
			log.Warnf("Failed to enumerate devices of card '%v': %v\n", card.String(), err)
			continue
		}
		cardInfo = append(cardInfo, fmt.Sprintf("%v  number=%v path=%v", card.Title, card.Number, card.Path))
		for _, dev := range cardDevices {
			device := &types.AudioDevice{
				Name:       fmt.Sprintf("%v [%v]", card.Title, dev.Title),
				Identifier: fmt.Sprintf("hw:%v,%v", card.Number, dev.Number),
				CanPlay:    dev.Play,
				CanRecord:  dev.Record,
			}
			results = append(results, device)
			cardInfo = append(cardInfo, fmt.Sprintf("\tNumber: %v", dev.Number))
			cardInfo = append(cardInfo, fmt.Sprintf("\tPath:   %v", dev.Path))
			cardInfo = append(cardInfo, fmt.Sprintf("\tPlay:   %v", dev.Play))
			cardInfo = append(cardInfo, fmt.Sprintf("\tRecord: %v", dev.Record))
			cardInfo = append(cardInfo, fmt.Sprintf("\tTitle:  %v", dev.Title))
			cardInfo = append(cardInfo, fmt.Sprintf("\tType:   %v", dev.Type))
		}
		log.Debugf("%v\n", strings.Join(cardInfo, "\n"))

	}
	return results, nil
}

// func ListDevicesWithALSAUtils() ([]*ALSADevice, error) {
// 	playOutput, err := exec.Command("aplay -l").Output()
// 	if err != nil {
// 		return nil, err
// 	}
// 	recordOutput, err := exec.Command("arecord -l").Output()
// 	if err != nil {
// 		return nil, err
// 	}

// }
