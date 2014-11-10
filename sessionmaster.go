package main

var sessions []SessionManager

func InitSessions() {
	var firstManager = SessionManager{}
	sessions = append(sessions, firstManager)
}

func ProcessMessage(m []byte) ([]byte, []byte) {

	return sessions[0].ProcessMessage(m)

}

func GetCurrentState() []Profile {

	return sessions[0].GetCurrentState()
}
