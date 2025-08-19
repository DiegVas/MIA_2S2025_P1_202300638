package structs

type UserParams struct {
	User string
	Pass string
	Grp  string
}

type RmUserParams struct {
	User string
}

type ChgrpParams struct {
	User string
	Grp  string
}
