package model

type CoverArt struct {
	ObjectName  string
	IsPrimary   bool    // takes precedence over volume when determining primary cover
	Volume      *string // unique per manga, except for null value which are allowed to have multiple entries
	Description *string

	staging bool
}

func NewCoverArt(objectName string, isPrimary bool, volume, description *string) (*CoverArt, error) {
	if err := validateVolume(volume); err != nil {
		return nil, err
	}

	return &CoverArt{
		ObjectName:  objectName,
		Volume:      volume,
		IsPrimary:   isPrimary,
		Description: description,
		staging:     false,
	}, nil
}

func NewStagingCoverArt(objectName string, isPrimary bool, volume, description *string) (*CoverArt, error) {
	if err := validateVolume(volume); err != nil {
		return nil, err
	}

	return &CoverArt{
		ObjectName:  objectName,
		Volume:      volume,
		IsPrimary:   isPrimary,
		Description: description,
		staging:     true,
	}, nil
}

func (c *CoverArt) ToStaged(objectName string) (*CoverArt, error) {
	return NewCoverArt(objectName, c.IsPrimary, c.Volume, c.Description)
}

func (c CoverArt) IsStaging() bool {
	return c.staging
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

		if cover.Volume != nil {
			v := *cover.Volume
			if err := validateVolume(&v); err != nil {
				return err
			}
			if uniqueVolumes[v] {
				return ErrVolumeAlreadyExists.WithArg("volume", v)
			}
			uniqueVolumes[v] = true
		}
	}

	return nil
}
