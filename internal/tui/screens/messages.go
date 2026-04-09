package screens

// ScreenName identifies which screen to navigate to.
type ScreenName int

const (
	ScreenOnboarding ScreenName = iota
	ScreenHome
	ScreenNewEntry
	ScreenBrowse
	ScreenSearch
	ScreenStats
	ScreenSettings
)

// ChangeScreenMsg tells the root App to switch to a new screen.
type ChangeScreenMsg struct {
	Screen ScreenName
	Data   interface{}
}
