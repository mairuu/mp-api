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
