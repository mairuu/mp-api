package model

// todo: make volume a nullable
type CoverArt struct {
	IsPrimary   bool   // takes precedence over volume when determining primary cover
	Volume      string // unique per manga, except for null/empty values which are allowed to have multiple entries
	ObjectName  string
	Description string
}

func NewCoverArt(volume, objectName, description string, isPrimary bool) (*CoverArt, error) {
	if err := validateVolume(&volume); err != nil {
		return nil, err
	}
	return &CoverArt{
		Volume:      volume,
		IsPrimary:   isPrimary,
		ObjectName:  objectName,
		Description: description,
	}, nil
}

func validateCoverArts(covers []CoverArt) error {
	foundPrimary := false
	uniqueVolumes := make(map[string]bool)

	for _, cover := range covers {
		if cover.IsPrimary {
			if foundPrimary {
				return ErrMultiplePrimaryCovers
			}
			foundPrimary = true
		}

		if cover.Volume == "" {
			continue
		}
		if err := validateVolume(&cover.Volume); err != nil {
			return err
		}
		if uniqueVolumes[cover.Volume] {
			return ErrVolumeAlreadyExists.WithArg("volume", cover.Volume)
		}
		uniqueVolumes[cover.Volume] = true
	}

	return nil
}
