package model

type ChapterPage struct {
	Width      int
	Height     int
	ObjectName string

	staging bool
}

func NewStagingChapterPage(objectName string) ChapterPage {
	return ChapterPage{
		ObjectName: objectName,
		Width:      0,
		Height:     0,
		staging:    true,
	}
}

func NewChapterPage(objectName string, width, height int) ChapterPage {
	return ChapterPage{
		ObjectName: objectName,
		Width:      width,
		Height:     height,
		staging:    false,
	}
}

func (p ChapterPage) IsStaging() bool {
	return p.staging
}

func validateChapterPage(page *ChapterPage) error {
	if page.ObjectName == "" {
		return ErrEmptyPageObjectName.WithMessage("page object name cannot be empty")
	}

	if !page.IsStaging() {
		if page.Width <= 0 {
			return ErrInvalidPageWidth.WithMessage("page width must be greater than zero")
		}
		if page.Height <= 0 {
			return ErrInvalidPageHeight.WithMessage("page height must be greater than zero")
		}
	}

	return nil
}
