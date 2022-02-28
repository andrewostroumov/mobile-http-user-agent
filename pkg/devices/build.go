package devices

import (
	"fmt"
	"github.com/andrewostroumov/mobile-http-user-agent/pkg/repo/types"
	"math/rand"
	"regexp"
	"strconv"
	"time"
)

const MinChromeRev = 64
const MinAndroidAPI = 27

var chromeMajorRegexp = regexp.MustCompile(`(?P<major>^\d+)`)

func BuildMobileAgent(device *types.Device) string {
	s := rand.NewSource(time.Now().Unix())
	r := rand.New(s)
	ch := buildChrome(r)
	an := buildAndroid(r, device)

	return fmt.Sprintf("Mozilla/5.0 (Linux; Android %v; %v) AppleWebKit/%v (KHTML, like Gecko) Chrome/%v Mobile Safari/%v", an.Version, device.Build, ch.WebkitRev, ch.ChromeRev, ch.WebkitRev)
}

func buildChrome(r *rand.Rand) Chrome {
	min := minChromeIndex()
	if min == 0 {
		return chr[len(chr)-1]
	}

	n := r.Intn(len(chr[min:]))
	return chr[min+n]
}

func buildAndroid(r *rand.Rand, device *types.Device) Android {
	min := minAndroidIndex()
	if min == 0 {
		return and[len(and)-1]
	}

	n := r.Intn(len(and[min:]))
	return and[min+n]
}

func minChromeIndex() int {
	for i, v := range chr {
		match := chromeMajorRegexp.FindStringSubmatch(v.ChromeRev)
		if len(match) == 0 {
			continue
		}

		fmt.Println(v)
		for j, name := range chromeMajorRegexp.SubexpNames() {
			if name == "major" {
				major, err := strconv.Atoi(match[j])
				if err != nil {
					continue
				}

				if major >= MinChromeRev {
					return i
				}
			}
		}
	}

	return 0
}

func minAndroidIndex() int {
	for i, v := range and {
		if v.API >= MinAndroidAPI {
			return i
		}
	}

	return 0
}
