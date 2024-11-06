package handler

type productResp struct {
	ID           int32  `json:"id" validate:"required,min=1"`
	Name         string `json:"name" validate:"required,min=1"`
	Description  string `json:"description"`
	Price        int32  `json:"price" validate:"min=0"`
	Availability int32  `json:"availability" validate:"min=0"`
}
