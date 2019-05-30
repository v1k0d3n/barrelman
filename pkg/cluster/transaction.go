package cluster

import (
	"github.com/charter-oss/structured/errors"
	"github.com/charter-oss/structured/log"
)

// Tiller Rollback
// Tiller keeps track of all information regarding a release in a configmap
// The release information is versioned and can be re-applied

// Barrelman Rollback
// The state is the combined release versions metadata needed to
// command Tiller to re-apply a given release version.

// State versioning
// States are versioned as well, and a version of the state can be applied to the system.
// Barrelman will command Tiller to re-apply each release version as necassary to make the
// current running release versions match the stored state.

// A transaction is the bounding ends of a state change.

// A transaction can be canceled which causes the in-progress state change to be rolled back
// to the previous state.

// Once a manifest has been applied with changes, and fully succeeeds, a new state will be recorded.
// Partially applied manifests will not result in a new state being recorded.

type NewTransactioner interface {
	NewTransaction(string) (*Transaction, error)
}

type Transactioner interface {
	WriteNewVersion() error
	Complete() error
	Versions() *Versions
	Cancel() error
	Canceled() bool
	Started() bool
	Completed() bool
}

type Transaction struct {
	ManifestName string
	startState   *State
	endState     *State
	canceled     bool
	changed      bool
	session      *Session
}

type State struct {
	Versions  *Versions
	completed bool
}

type ChangedRelease struct {
	ReleaseName     string
	OriginalVersion int32
	NewVersion      int32
}

// NewTransaction initializes a transaction data structure
// the Startstate field
// this transaction can then be used to track changes and perform rollbacks
func (s *Session) NewTransaction(manifestName string) (*Transaction, error) {
	currentVersions, err := s.GetVersions(manifestName)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get current versions while creating rollback transaction")
	}
	transaction := &Transaction{
		ManifestName: manifestName,
		session:      s,
		startState: &State{
			Versions: currentVersions,
		},
		endState: &State{
			Versions: NewVersions(manifestName),
		},
	}

	if err := transaction.startTransaction(); err != nil {
		return nil, errors.Wrap(err, "failed to create new rollback transaction")
	}
	return transaction, nil
}

// startTransaction populates startState with the currently running versions
func (t *Transaction) startTransaction() error {
	if t.Started() {
		return errors.WithFields(errors.Fields{
			"ManifestName": t.ManifestName,
		}).New("cannot start this transaction, its already been started")
	}
	if t.Completed() {
		return errors.WithFields(errors.Fields{
			"ManifestName": t.ManifestName,
		}).New("cannot start this transaction, its already been completed")
	}
	versions, err := t.session.GetVersions(t.ManifestName)
	if err != nil {
		return errors.Wrap(err, "failed to start worllback transaction")
	}
	t.startState.Versions = versions
	t.startState.completed = true
	return nil
}

func (t *Transaction) completeTransaction() error {
	if !t.Started() {
		return errors.WithFields(errors.Fields{
			"ManifestName": t.ManifestName,
		}).New("cannot complete this transaction, it hasent been started")
	}
	if t.Completed() {
		return errors.WithFields(errors.Fields{
			"ManifestName": t.ManifestName,
		}).New("cannot complete this transaction, its already been completed")
	}
	//TODO: Calculate differences

	t.endState.completed = true // we do not attempt this twice
	changedList, changed := t.calculateChanged()
	if changed {
		//Log the changed releases
		for _, v := range changedList {
			log.WithFields(log.Fields{
				"ReleaseName":     v.ReleaseName,
				"OriginalVersion": v.OriginalVersion,
				"NewVersion":      v.NewVersion,
			}).Debug("Release was changed")
		}
		//Write all releases to new new version
		if err := t.WriteNewVersion(); err != nil {
			return errors.Wrap(err, "Failed to write new rollback state")
		}
	} else {
		//No changes, no new version
		log.Debug("No change")
	}
	return nil
}

func (t *Transaction) WriteNewVersion() error {
	return t.session.WriteVersions(t.Versions())
}

// SetChanged sets a global transaction changed flag allowing it to generate a new version
func (t *Transaction) SetChanged() {
	t.changed = true
}

// Changed returns the global transaction changed flag allowing it to generate a new version
func (t *Transaction) Changed() bool {
	return t.changed
}

// Complete populates endState with the currently running versions
// then writes a new versioned release
func (t *Transaction) Complete() error {
	return t.completeTransaction()
}

// Versions returns the endstate Versions
func (t *Transaction) Versions() *Versions {
	return t.endState.Versions
}

// Cancel sets the versions and installation state to the previously recorded versions
// effectivly undoing any commanded actions performed within the transaction
func (t *Transaction) Cancel() error {
	if t.Canceled() {
		//Tollerate multiple calls to Cancel
		return nil
	}
	if t.Completed() {
		//Tollerate canceled after completed
		return nil
	}
	if !t.Started() {
		return errors.WithFields(errors.Fields{
			"ManifestName": t.ManifestName,
		}).New("cannot cancel this transaction, it hasent been started")
	}
	t.canceled = true
	return nil
}

//Close

func (t *Transaction) Canceled() bool {
	return t.canceled
}

func (t *Transaction) Started() bool {
	return t.startState.completed
}

func (t *Transaction) Completed() bool {
	return t.endState.completed
}

func (t *Transaction) calculateChanged() ([]*ChangedRelease, bool) {

	changedReleases := []*ChangedRelease{}
	for _, version := range t.endState.Versions.Data {
		if version.IsModified() {
			changedReleases = append(changedReleases, &ChangedRelease{
				ReleaseName: version.Name,
				NewVersion:  version.Revision,
			})
		}
	}

	return changedReleases, len(changedReleases) > 0 || t.changed
}

func (state *State) Completed() bool {
	return state.completed
}
