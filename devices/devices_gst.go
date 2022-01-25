package devices

import (
	"bufio"
	"bytes"
	"os/exec"
	"regexp"
	"strings"

	"github.com/gurupras/dhwani_backend_p2p/types"
	"github.com/gurupras/dhwani_backend_p2p/utils"
)

func GetAudioDevices() ([]*types.AudioDevice, error) {
	stdout, err := exec.Command("gst-device-monitor-1.0").Output()
	if err != nil {
		return nil, err
	}
	return parseGstDeviceMonitorOutput(stdout)
}

func parseGstDeviceMonitorOutput(b []byte) ([]*types.AudioDevice, error) {
	results := make([]*types.AudioDevice, 0)
	scanner := bufio.NewScanner(bytes.NewBuffer(b))
	scanner.Split(bufio.ScanLines)

	deviceFoundRegex := regexp.MustCompile(`\s*Device\s+found:.*`)
	nameRegex := regexp.MustCompile(`\s+name\s+:(?P<name>.*)`)
	sinkClassRegex := regexp.MustCompile(`\s+class\s+:\s+Audio/Sink`)
	sourceClassRegex := regexp.MustCompile(`\s+class\s+:\s+Audio/Source`)
	deviceRegex := regexp.MustCompile(`\s+gst-launch.* device=(?P<device>\S+).*`)
	for scanner.Scan() {
		line := scanner.Text()
		if deviceFoundRegex.MatchString(line) {
			name := ""
			canRecord := false
			canPlay := false
			identifier := ""
			for scanner.Scan() {
				line := scanner.Text()

				if nameRegex.MatchString(line) {
					m := utils.GetRegexGroups(nameRegex, line)
					name = strings.TrimSpace(m["name"])
				} else if sinkClassRegex.MatchString(line) {
					canPlay = true
				} else if sourceClassRegex.MatchString(line) {
					canRecord = true
				}
				if strings.Index(line, "gst-launch") > 0 {
					m := utils.GetRegexGroups(deviceRegex, line)
					identifier = strings.TrimSpace(m["device"])
					results = append(results, &types.AudioDevice{
						Name:       name,
						Identifier: identifier,
						CanPlay:    canPlay,
						CanRecord:  canRecord,
					})
					break
				}
			}
		}
	}
	return results, nil
}
