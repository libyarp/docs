package contacts

import "context"

func RegisterContact() {
	c := NewContactsServiceClient("localhost:9027")
	contact := &Contact{
		ID:      nil,
		Name:    "Paul",
		Surname: "Appleseed",
		Company: nil,
		Emails:  []string{"paul.appleseed@icloud.com"},
	}
	_, err := c.UpsertContact(context.Background(), contact, nil)
	if err != nil {
		// ...
	}
}
