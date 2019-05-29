package barrelman

// TransitionState defines for computed action used in Apply, Rollback
type TransitionState int

const (
	// NoChange means no differences were detected between the proposed and running releases
	NoChange TransitionState = iota
	// Installable means there is no running release so an Install will be performed
	Installable
	// Upgradable indicates there is a running release that will be upgraded
	Upgradable
	// Replaceable means that the conditions were met to perform a "force" upgrade (delete/install)
	Replaceable
	// Deletable means the target release will be deleted
	Deletable
	// Undeleteable is a situion where the release has been deleted but not purged
	// and we want to install again. The solution is to rollback to the last revision first
	// then Upgrade
	Undeleteable
)
