package dto

type CreateStudioRequest struct {
	Name          string `json:"name" validate:"required,min=1,max=100"`
	Seat_Capacity int    `json:"seat_capacity" validate:"required,min=1,max=600"`
	Location      string `json:"location" validate:"required,min=1"`
}

type UpdateStudioRequest struct {
	Name          *string `json:"name,omitempty" validate:"omitempty,min=1,max=100"`
	Seat_Capacity *int    `json:"seat_capacity,omitempty" validate:"omitempty,min=1,max=600"`
	Location      string  `json:"location,omitempty" validate:"omitempty,min=1"`
}

