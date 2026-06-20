package models

type PlatformType int

const (
	PlatformAll         PlatformType = -1
	PlatformSteam       PlatformType = 0
	PlatformOculus      PlatformType = 1
	PlatformPlayStation PlatformType = 2
	PlatformMicrosoft   PlatformType = 3
	PlatformHeadlessBot PlatformType = 4
	PlatformIOS         PlatformType = 5
)

func (p PlatformType) IsKnown() bool {
	return p >= PlatformSteam && p <= PlatformIOS
}

func (p PlatformType) String() string {
	switch p {
	case PlatformSteam:
		return "Steam"
	case PlatformOculus:
		return "Oculus"
	case PlatformPlayStation:
		return "PlayStation"
	case PlatformMicrosoft:
		return "Microsoft"
	case PlatformHeadlessBot:
		return "HeadlessBot"
	case PlatformIOS:
		return "iOS"
	default:
		return "Unknown"
	}
}
