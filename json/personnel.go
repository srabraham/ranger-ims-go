package json

type Personnel []Person
type Person struct {
	Handle      string `json:"handle"`
	Status      string `json:"status"`
	Onsite      bool   `json:"onsite"`
	DirectoryID int64  `json:"directory_id"`
}
