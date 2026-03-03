package model

type ChapterPage struct {
	Width      int
	Height     int
	ObjectName string
}

func validateChapterPage(page *ChapterPage) error {
	if page.Width <= 0 {
		return ErrInvalidPageWidth.WithMessage("page width must be greater than zero")
	}
	if page.Height <= 0 {
		return ErrInvalidPageHeight.WithMessage("page height must be greater than zero")
	}
	if page.ObjectName == "" {
		return ErrEmptyPageObjectName.WithMessage("page object name cannot be empty")
	}
	return nil
}
