package service

type UploadDTO struct {
	AcceptedFiles []AcceptedFile `json:"accepts"`
	RejectedFiles []RejectedFile `json:"rejects"`
}

type AcceptedFile struct {
	OriginalFileName string `json:"original_file_name"`
	ObjectName       string `json:"object_name"`
}

type RejectedFile struct {
	OriginalFileName string `json:"original_file_name"`
	Error            string `json:"error"`
}
