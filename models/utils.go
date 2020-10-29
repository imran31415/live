package models

func boolFromBoolPointer(b *bool) bool {
	if b != nil && *b == true {
		return true
	}
	return false
}

func stringFromStringPointer(s *string) string {
	if s == nil {
		return ""
	} else {
		return *s
	}
}

func int64FromInt64Pointer(i *int64) int64 {
	if i == nil {
		return 0
	} else {
		return *i
	}
}
