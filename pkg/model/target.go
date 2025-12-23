package model

// TargetDevice enumerates supported hardware targets.
type TargetDevice string

const (
	TargetSD2IEC      TargetDevice = "sd2iec"
	TargetPi1541      TargetDevice = "pi1541"
	TargetKungFuFlash TargetDevice = "kungfuflash"
	TargetUltimate    TargetDevice = "ultimate"
)

// TargetProfile centralizes hardware constraints.
type TargetProfile struct {
	Target         TargetDevice
	MaxNameLen     int
	DisplayNameLen int
	Notes          string
}

// ProfileFor returns the constraints for a target device.
func ProfileFor(target TargetDevice) TargetProfile {
	switch target {
	case TargetSD2IEC:
		return TargetProfile{Target: TargetSD2IEC, MaxNameLen: 16, Notes: "Commodore DOS filename length"}
	case TargetPi1541:
		return TargetProfile{Target: TargetPi1541, MaxNameLen: 16, Notes: "Behaves like 1541/Commodore DOS"}
	case TargetKungFuFlash:
		return TargetProfile{Target: TargetKungFuFlash, MaxNameLen: 255, DisplayNameLen: 32, Notes: "Menu display truncates around 32 chars"}
	case TargetUltimate:
		return TargetProfile{Target: TargetUltimate, MaxNameLen: 255, Notes: "Filesystem long filename typical maximum"}
	default:
		return TargetProfile{Target: target, MaxNameLen: 16}
	}
}
