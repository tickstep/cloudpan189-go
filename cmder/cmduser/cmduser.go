package cmduser

type CmdUser struct {
	UID  uint64
	Name string
	Workdir string
}

type CmdUserList []*CmdUser
