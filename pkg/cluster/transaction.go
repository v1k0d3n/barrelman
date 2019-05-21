package cluster

import (
	"github.com/charter-oss/structured/errors"
)

//Tiller Rollback
// Tiller keeps track of all information regarding a release in a configmap
// The release information is versioned and can be re-applied

//Barrelman Rollback
// The state is the combined release versions metadata needed to
// command Tiller to re-apply a given release version.

//State versioning
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

type Transaction struct {
	ManifestName string
	startState   *State
	endState     *State
	canceled     bool
	session      *Session
}

type State struct {
	Versions  *Versions
	completed bool
}

//NewTransaction initializes a transaction data structure
// the Startstate field
// this transaction can then be used to track changes and perform rollbacks
func (s *Session) NewTransaction(manifestName string) (*Transaction, error) {
	transaction := &Transaction{
		ManifestName: manifestName,
		session:      s,
		startState: &State{
			Versions: NewVersions(manifestName),
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

//startTransaction populates startState with the currently running versions
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

//completeTransaction populates endState with the currently running versions
// then writes a new versioned release
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
	versions, err := t.session.GetVersions(t.ManifestName)
	if err != nil {
		return errors.Wrap(err, "failed to start rollback transaction")
	}
	t.endState.Versions = versions
	t.endState.completed = true // we do not attempt this twice
	if err := t.session.WriteVersions(versions); err != nil {
		return errors.Wrap(err, "Failed to write new rollback state")
	}
	return nil
}

//Complete
func (t *Transaction) Complete() error {
	return t.completeTransaction()
}

//Cancel sets the versions and installation state to the previously recorded versions
// effectivly undoing any commanded actions performed within the transaction
func (t *Transaction) Cancel() error {
	if t.Canceled() {
		//Tollerate multiple calls to Cancel
		return nil
	}
	if !t.Started() {
		return errors.WithFields(errors.Fields{
			"ManifestName": t.ManifestName,
		}).New("cannot cancel this transaction, it hasent been started")
	}
	if t.Completed() {
		return errors.WithFields(errors.Fields{
			"ManifestName": t.ManifestName,
		}).New("cannot cancel this transaction, its already been completed")
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

func (state *State) Completed() bool {
	return state.completed
}
