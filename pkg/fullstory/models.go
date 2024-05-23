package fullstory

type User struct {
	ID             string `json:"id"`
	UID            string `json:"uid"`
	Name           string `json:"display_name"`
	Email          string `json:"email"`
	IsBeingDeleted bool   `json:"is_being_deleted"`
}
