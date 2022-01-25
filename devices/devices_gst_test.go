package devices

import (
	"testing"

	"github.com/gurupras/dhwani_backend_p2p/types"
	"github.com/stretchr/testify/require"
)

func TestParseGstDeviceMonitorOutput(t *testing.T) {
	require := require.New(t)
	rawStr := []byte(`Probing devices...


Device found:

	name  : Monitor of Starship/Matisse HD Audio Controller Digital Stereo (IEC958)
	class : Audio/Source
	caps  : audio/x-raw, format=(string){ S16LE, S16BE, F32LE, F32BE, S32LE, S32BE, S24LE, S24BE, S24_32LE, S24_32BE, U8 }, layout=(string)interleaved, rate=(int)[ 1, 384000 ], channels=(int)[ 1, 32 ];
	        audio/x-alaw, rate=(int)[ 1, 384000 ], channels=(int)[ 1, 32 ];
	        audio/x-mulaw, rate=(int)[ 1, 384000 ], channels=(int)[ 1, 32 ];
	properties:
		device.description = "Monitor\ of\ Starship/Matisse\ HD\ Audio\ Controller\ Digital\ Stereo\ \(IEC958\)"
		device.class = monitor
		alsa.card = 1
		alsa.card_name = "HD-Audio\ Generic"
		alsa.long_card_name = "HD-Audio\ Generic\ at\ 0xfcb00000\ irq\ 66"
		alsa.driver_name = snd_hda_intel
		device.bus_path = pci-0000:0a:00.4
		sysfs.path = /devices/pci0000:00/0000:00:08.1/0000:0a:00.4/sound/card1
		device.bus = pci
		device.vendor.id = 1022
		device.vendor.name = "Advanced\ Micro\ Devices\,\ Inc.\ \[AMD\]"
		device.product.id = 1487
		device.product.name = "Starship/Matisse\ HD\ Audio\ Controller"
		device.string = 1
		module-udev-detect.discovered = 1
		device.icon_name = audio-card-pci
		is-default = true
	gst-launch-1.0 pulsesrc device=alsa_output.pci-0000_0a_00.4.iec958-stereo.monitor ! ...


Device found:

	name  : Monitor of Vega 10 HDMI Audio [Radeon Vega 56/64] Digital Stereo (HDMI 6)
	class : Audio/Source
	caps  : audio/x-raw, format=(string){ S16LE, S16BE, F32LE, F32BE, S32LE, S32BE, S24LE, S24BE, S24_32LE, S24_32BE, U8 }, layout=(string)interleaved, rate=(int)[ 1, 384000 ], channels=(int)[ 1, 32 ];
	        audio/x-alaw, rate=(int)[ 1, 384000 ], channels=(int)[ 1, 32 ];
	        audio/x-mulaw, rate=(int)[ 1, 384000 ], channels=(int)[ 1, 32 ];
	properties:
		device.description = "Monitor\ of\ Vega\ 10\ HDMI\ Audio\ \[Radeon\ Vega\ 56/64\]\ Digital\ Stereo\ \(HDMI\ 6\)"
		device.class = monitor
		alsa.card = 0
		alsa.card_name = "HDA\ ATI\ HDMI"
		alsa.long_card_name = "HDA\ ATI\ HDMI\ at\ 0xfcca0000\ irq\ 64"
		alsa.driver_name = snd_hda_intel
		device.bus_path = pci-0000:08:00.1
		sysfs.path = /devices/pci0000:00/0000:00:03.1/0000:06:00.0/0000:07:00.0/0000:08:00.1/sound/card0
		device.bus = pci
		device.vendor.id = 1002
		device.vendor.name = "Advanced\ Micro\ Devices\,\ Inc.\ \[AMD/ATI\]"
		device.product.id = aaf8
		device.product.name = "Vega\ 10\ HDMI\ Audio\ \[Radeon\ Vega\ 56/64\]"
		device.string = 0
		module-udev-detect.discovered = 1
		device.icon_name = audio-card-pci
		is-default = false
	gst-launch-1.0 pulsesrc device=alsa_output.pci-0000_08_00.1.hdmi-stereo-extra5.monitor ! ...


Device found:

	name  : Starship/Matisse HD Audio Controller Digital Stereo (IEC958)
	class : Audio/Sink
	caps  : audio/x-raw, format=(string){ S16LE, S16BE, F32LE, F32BE, S32LE, S32BE, S24LE, S24BE, S24_32LE, S24_32BE, U8 }, layout=(string)interleaved, rate=(int)[ 1, 384000 ], channels=(int)[ 1, 32 ];
	        audio/x-alaw, rate=(int)[ 1, 384000 ], channels=(int)[ 1, 32 ];
	        audio/x-mulaw, rate=(int)[ 1, 384000 ], channels=(int)[ 1, 32 ];
	properties:
		alsa.resolution_bits = 16
		device.api = alsa
		device.class = sound
		alsa.class = generic
		alsa.subclass = generic-mix
		alsa.name = "ALC887-VD\ Digital"
		alsa.id = "ALC887-VD\ Digital"
		alsa.subdevice = 0
		alsa.subdevice_name = "subdevice\ \#0"
		alsa.device = 1
		alsa.card = 1
		alsa.card_name = "HD-Audio\ Generic"
		alsa.long_card_name = "HD-Audio\ Generic\ at\ 0xfcb00000\ irq\ 66"
		alsa.driver_name = snd_hda_intel
		device.bus_path = pci-0000:0a:00.4
		sysfs.path = /devices/pci0000:00/0000:00:08.1/0000:0a:00.4/sound/card1
		device.bus = pci
		device.vendor.id = 1022
		device.vendor.name = "Advanced\ Micro\ Devices\,\ Inc.\ \[AMD\]"
		device.product.id = 1487
		device.product.name = "Starship/Matisse\ HD\ Audio\ Controller"
		device.string = iec958:1
		device.buffering.buffer_size = 352768
		device.buffering.fragment_size = 176384
		device.access_mode = mmap+timer
		device.profile.name = iec958-stereo
		device.profile.description = "Digital\ Stereo\ \(IEC958\)"
		device.description = "Starship/Matisse\ HD\ Audio\ Controller\ Digital\ Stereo\ \(IEC958\)"
		module-udev-detect.discovered = 1
		device.icon_name = audio-card-pci
		is-default = true
	gst-launch-1.0 ... ! pulsesink device=alsa_output.pci-0000_0a_00.4.iec958-stereo


Device found:

	name  : Vega 10 HDMI Audio [Radeon Vega 56/64] Digital Stereo (HDMI 6)
	class : Audio/Sink
	caps  : audio/x-raw, format=(string){ S16LE, S16BE, F32LE, F32BE, S32LE, S32BE, S24LE, S24BE, S24_32LE, S24_32BE, U8 }, layout=(string)interleaved, rate=(int)[ 1, 384000 ], channels=(int)[ 1, 32 ];
	        audio/x-alaw, rate=(int)[ 1, 384000 ], channels=(int)[ 1, 32 ];
	        audio/x-mulaw, rate=(int)[ 1, 384000 ], channels=(int)[ 1, 32 ];
	properties:
		alsa.resolution_bits = 16
		device.api = alsa
		device.class = sound
		alsa.class = generic
		alsa.subclass = generic-mix
		alsa.name = "HDMI\ 5"
		alsa.id = "HDMI\ 5"
		alsa.subdevice = 0
		alsa.subdevice_name = "subdevice\ \#0"
		alsa.device = 11
		alsa.card = 0
		alsa.card_name = "HDA\ ATI\ HDMI"
		alsa.long_card_name = "HDA\ ATI\ HDMI\ at\ 0xfcca0000\ irq\ 64"
		alsa.driver_name = snd_hda_intel
		device.bus_path = pci-0000:08:00.1
		sysfs.path = /devices/pci0000:00/0000:00:03.1/0000:06:00.0/0000:07:00.0/0000:08:00.1/sound/card0
		device.bus = pci
		device.vendor.id = 1002
		device.vendor.name = "Advanced\ Micro\ Devices\,\ Inc.\ \[AMD/ATI\]"
		device.product.id = aaf8
		device.product.name = "Vega\ 10\ HDMI\ Audio\ \[Radeon\ Vega\ 56/64\]"
		device.string = "hdmi:0\,5"
		device.buffering.buffer_size = 352768
		device.buffering.fragment_size = 176384
		device.access_mode = mmap+timer
		device.profile.name = hdmi-stereo-extra5
		device.profile.description = "Digital\ Stereo\ \(HDMI\ 6\)"
		device.description = "Vega\ 10\ HDMI\ Audio\ \[Radeon\ Vega\ 56/64\]\ Digital\ Stereo\ \(HDMI\ 6\)"
		module-udev-detect.discovered = 1
		device.icon_name = audio-card-pci
		is-default = false
	gst-launch-1.0 ... ! pulsesink device=alsa_output.pci-0000_08_00.1.hdmi-stereo-extra5

`)

	expected := []*types.AudioDevice{
		&types.AudioDevice{
			Name:       "Monitor of Starship/Matisse HD Audio Controller Digital Stereo (IEC958)",
			Identifier: "alsa_output.pci-0000_0a_00.4.iec958-stereo.monitor",
			CanPlay:    false,
			CanRecord:  true,
		},
		&types.AudioDevice{
			Name:       "Monitor of Vega 10 HDMI Audio [Radeon Vega 56/64] Digital Stereo (HDMI 6)",
			Identifier: "alsa_output.pci-0000_08_00.1.hdmi-stereo-extra5.monitor",
			CanPlay:    false,
			CanRecord:  true,
		},
		&types.AudioDevice{
			Name:       "Starship/Matisse HD Audio Controller Digital Stereo (IEC958)",
			Identifier: "alsa_output.pci-0000_0a_00.4.iec958-stereo",
			CanPlay:    true,
			CanRecord:  false,
		},
		&types.AudioDevice{
			Name:       "Vega 10 HDMI Audio [Radeon Vega 56/64] Digital Stereo (HDMI 6)",
			Identifier: "alsa_output.pci-0000_08_00.1.hdmi-stereo-extra5",
			CanPlay:    true,
			CanRecord:  false,
		},
	}

	devices, err := parseGstDeviceMonitorOutput(rawStr)
	require.Nil(err)

	require.Equal(len(expected), len(devices))
	for idx, exp := range expected {
		got := devices[idx]
		require.Equal(exp, got)
	}
}
