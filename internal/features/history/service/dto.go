package service

type RecentReadDTO struct {
	MangaID         string  `json:"manga_id"`
	MangaTitle      string  `json:"manga_title"`
	CoverObjectName *string `json:"cover_object_name"`
	ChapterID       string  `json:"chapter_id"`
	ChapterTitle    string  `json:"chapter_title"`
	ChapterNumber   string  `json:"chapter_number"`
	Progress        float32 `json:"progress"`
	ReadAt          string  `json:"read_at"`
}

type MangaReadDTO struct {
	ChapterID string  `json:"chapter_id"`
	Progress  float32 `json:"progress"`
	ReadAt    string  `json:"read_at"`
}

type ChapterProgressDTO struct {
	ChapterID string  `json:"id" binding:"required,uuid"`
	Progress  float32 `json:"progress" binding:"gte=0,lte=1"`
}

type MarkChaptersAsReadDTO struct {
	Chapters []ChapterProgressDTO `json:"chapters" binding:"required,dive"`
}

type RemoveReadChapter struct {
	ChapterID string `json:"id" binding:"required,uuid"`
}

type UnmarkChaptersAsReadDTO struct {
	Chapters []RemoveReadChapter `json:"chapters" binding:"required,dive"`
}
