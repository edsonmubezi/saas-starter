package user

func SanitizeUsers(users []User) {
	for i := range users {
		users[i].Password = ""
	}
}
