package service

type UploadDTO struct {
	AcceptedFiles []AcceptedFile `json:"accepts"`
	RejectedFiles []RejectedFile `json:"rejects"`
}

type AcceptedFile struct {
	RefID            *string `json:"ref_id,omitempty"`
	OriginalFileName string  `json:"original_file_name"`
	ObjectName       string  `json:"object_name"`
}

type RejectedFile struct {
	RefID            *string `json:"ref_id,omitempty"`
	OriginalFileName string  `json:"original_file_name"`
	Error            string  `json:"error"` // todo: change to reason code and message
}
