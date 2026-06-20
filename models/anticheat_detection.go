package models

type DetectionType int

const (
	DetectionMaliciousModule DetectionType = iota
	DetectionMemoryTamper
	DetectionManualMap
	DetectionDebugDetected
)

var detectionTypeNames = map[string]DetectionType{
	"malicious_module": DetectionMaliciousModule,
	"memory_tamper":    DetectionMemoryTamper,
	"manual_map":       DetectionManualMap,
	"debug_detected":   DetectionDebugDetected,
}

func ParseDetectionType(s string) (DetectionType, bool) {
	t, ok := detectionTypeNames[s]
	return t, ok
}

func (t DetectionType) ShouldClose() bool {
	switch t {
	case DetectionMaliciousModule, DetectionMemoryTamper, DetectionManualMap, DetectionDebugDetected:
		return true
	default:
		return false
	}
}
