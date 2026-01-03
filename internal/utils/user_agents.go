package utils

import (
	"fmt"

	"github.com/avct/uasurfer"
)

func UserAgentVersionToString(v uasurfer.Version) string {
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}
