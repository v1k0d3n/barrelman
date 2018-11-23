package sourcetype

import (
	"fmt"

	"github.com/charter-se/structured/errors"
)

const (
	Unknown = iota
	Missing = iota
	Git     = iota
	Local   = iota
)

type SourceType int

func Parse(s string) (SourceType, error) {
	switch s {
	case "":
		return SourceType(Missing), nil
	case "git":
		return SourceType(Git), nil
	case "local":
		return SourceType(Local), nil
	}
	return SourceType(Unknown), errors.WithFields(errors.Fields{"Type": s}).New("Failed to parse sourcetype")
}

func Print(st SourceType) string {
	switch st {
	case Unknown:
		return "unknown"
	case Git:
		return "git"
	case Local:
		return "local"
	default:
		return fmt.Sprintf("Unknown type: %v", st)
	}
}
