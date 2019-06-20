package barrelman

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/charter-oss/barrelman/pkg/cluster"
	"github.com/charter-oss/barrelman/pkg/manifest"
	"github.com/charter-oss/barrelman/pkg/version"
	"github.com/charter-oss/structured/errors"
	"github.com/charter-oss/structured/log"
)

const (
	Installable = iota
	Upgradable
	Replaceable
	Deletable
	NoChange
)

type ApplyCmd struct {
	Options    *CmdOptions
	Config     *Config
	LogOptions *[]string
}

type ReleaseTarget struct {
	ReleaseMeta *cluster.ReleaseMeta
	Chart       *cluster.Chart
	State       int
	Diff        []byte
	Changed     bool
}

type ReleaseTargets []*ReleaseTarget

type ReleaseGroup struct {
	Name      string
	Desc      string
	Sequenced bool
	Data      ReleaseTargets
}

type ReleaseGroups []*ReleaseGroup

type Control struct {
	ctx    context.Context
	cancel context.CancelFunc
	C      chan *ProcessResp
	wg     sync.WaitGroup
}

type ProcessResp struct {
	err error
}

func (cmd *ApplyCmd) Run(session cluster.Sessioner) error {
	var err error

	ver := version.Get()
	log.WithFields(log.Fields{
		"Version": ver.Version,
		"Commit":  ver.Commit,
		"Branch":  ver.Branch,
	}).Info("Barrelman")

	cmd.Config, err = GetConfigFromFile(cmd.Options.ConfigFile)
	if err != nil {
		return errors.Wrap(err, "got error while loading config")
	}

	if err := ensureWorkDir(cmd.Options.DataDir); err != nil {
		return errors.Wrap(err, "failed to create working directory")
	}

	log.Debug("connecting to cluster")
	if err = session.Init(); err != nil {
		return errors.Wrap(err, "failed to create new cluster session")
	}

	if session.GetKubeConfig() != "" {
		log.WithFields(log.Fields{
			"file": session.GetKubeConfig(),
		}).Info("Using kube config")
	}
	if session.GetKubeContext() != "" {
		log.WithFields(log.Fields{
			"file": session.GetKubeContext(),
		}).Info("Using kube context")
	}

	archiveGroups, err := processManifest(&manifest.Config{
		DataDir:      cmd.Options.DataDir,
		ManifestFile: cmd.Options.ManifestFile,
		AccountTable: cmd.Config.Account,
	}, cmd.Options.NoSync)
	if err != nil {
		return errors.Wrap(err, "apply failed")
	}

	releases, err := session.Releases()
	if err != nil {
		return errors.Wrap(err, "failed to get current releases")
	}

	rgs, err := cmd.ComputeReleases(session, archiveGroups, releases)
	if err != nil {
		return err
	}

	if err := rgs.dryRun(session); err != nil {
		return err
	}
	if cmd.Options.DryRun {
		log.Info("No errors")
		return nil
	}

	_, err = rgs.Diff(session)
	if err != nil {
		return err
	}
	if cmd.Options.Diff {
		rgs.LogDiff()
		return nil
	}
	err = rgs.Apply(session, cmd.Options)
	if err != nil {
		return errors.Wrap(err, "Manifest upgrade failed")
	}

	return nil
}

//IsReplaceable checks a release against the --force flag values to see if an existing release should be replaced via delete
func (cmd *ApplyCmd) isInForce(rel *cluster.ReleaseMeta) bool {
	//Checks for releases configured for Force by cmdline
	for _, r := range *cmd.Options.Force {
		if r == rel.MetaName || r == rel.ChartName || r == rel.ReleaseName {
			return true
		}
	}
	return false
}

//ComputeReleases configures each potential release with a current state
//states may be one of 'Installable', 'Upgradeable', or 'Replaceable'
func (cmd *ApplyCmd) ComputeReleases(
	session cluster.Sessioner,
	archiveGroups manifest.ArchiveGroups,
	currentReleases map[string]*cluster.ReleaseMeta) (ReleaseGroups, error) {
	rgs := []*ReleaseGroup{}

	for _, archiveGroup := range archiveGroups {
		releaseGroup := &ReleaseGroup{
			Name:      archiveGroup.Name,
			Desc:      archiveGroup.Desc,
			Sequenced: archiveGroup.Sequenced,
			Data:      ReleaseTargets{},
		}
		for _, v := range archiveGroup.ArchiveFiles.List {
			inChart, err := session.ChartFromArchive(v.Reader)
			if err != nil {
				return nil, errors.Wrap(err, "Failed to generate chart from archive")
			}

			// New release target from archive, Installable by default, may be modified later
			releaseTarget := &ReleaseTarget{
				State: Installable,
				ReleaseMeta: &cluster.ReleaseMeta{
					Chart:          inChart,
					ReleaseName:    v.ReleaseName,
					Namespace:      v.Namespace,
					ValueOverrides: v.Overrides,
					InstallWait:    v.InstallWait,
				},
			}

			// Check all current releases, change transition state to match
			for _, rel := range currentReleases {
				if rel.ReleaseName == v.ReleaseName {
					if cmd.isInForce(rel) || rel.Status == cluster.Status_FAILED {
						releaseTarget.State = Replaceable
					} else {
						releaseTarget.State = Upgradable
					}
				}
			}
			releaseGroup.Data = append(releaseGroup.Data, releaseTarget)
		}
		rgs = append(rgs, releaseGroup)
	}
	return rgs, nil
}

func (rgs ReleaseGroups) dryRun(session cluster.Sessioner) error {
	for _, releaseGroup := range rgs {
		for _, v := range releaseGroup.Data {
			v.ReleaseMeta.DryRun = true
			switch v.State {
			case Installable:
				_, _, err := session.InstallRelease(v.ReleaseMeta)
				if err != nil {
					return err
				}
			case Upgradable:
				_, err := session.UpgradeRelease(v.ReleaseMeta)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (rgs ReleaseGroups) Diff(session cluster.Sessioner) (ReleaseGroups, error) {
	var err error
	for _, releaseGroup := range rgs {
		for _, v := range releaseGroup.Data {
			v.ReleaseMeta.DryRun = true
			switch v.State {
			case Upgradable:
				v.Changed, v.Diff, err = session.DiffRelease(v.ReleaseMeta)
				if err != nil {
					return nil, err
				}
			}
		}
	}
	return rgs, nil
}

func (rgs ReleaseGroups) LogDiff() {
	for _, releaseGroup := range rgs {
		for _, v := range releaseGroup.Data {
			switch v.State {
			case Installable:
				log.WithFields(log.Fields{
					"Name":      v.ReleaseMeta.ReleaseName,
					"Namespace": v.ReleaseMeta.Namespace,
				}).Info("Would install")
			case Upgradable:
				if v.Changed {
					log.WithFields(log.Fields{
						"Name": v.ReleaseMeta.ReleaseName,
					}).Info("Diff")
					//Print the byte content and keep formatting, its fancy
					fmt.Printf("----%v\n%v_____\n", v.ReleaseMeta.MetaName, string(v.Diff))
				} else {
					log.WithFields(log.Fields{
						"Name": v.ReleaseMeta.ReleaseName,
					}).Info("No change")
				}
			}
		}
	}
}

// Apply performs actions on ReleaseTargets using pre-calculated transition states
func (rgs ReleaseGroups) Apply(session cluster.Sessioner, opt *CmdOptions) error {
	for _, releaseGroup := range rgs {
		log.WithFields(log.Fields{
			"Name": releaseGroup.Name,
			"Desc": releaseGroup.Desc,
		}).Debug("Processing Release Group")
		if err := releaseGroup.process(session, opt); err != nil {
			return err
		}
	}
	return nil
}

func (rg ReleaseGroup) process(session cluster.Sessioner, opt *CmdOptions) error {
	return rg.Data.process(session, opt, rg.Sequenced)
}

func (rt ReleaseTargets) process(session cluster.Sessioner, opt *CmdOptions, sequenced bool) error {
	control := NewControl(context.Background())

	for _, releaseTarget := range rt {
		if control.Canceled() {
			return nil
		}
		log.WithFields(log.Fields{
			"Name": releaseTarget.ReleaseMeta.ReleaseName,
		}).Warn("processing release target")
		releaseTarget.ReleaseMeta.DryRun = false
		releaseTarget.ReleaseMeta.InstallTimeout = 120
		go func(control *Control, v *ReleaseTarget) {
			control.Add(1)
			defer control.Done()
			switch v.State {
			case Installable, Replaceable:
				if err := func() error {
					//This closure removes a "break OUT"
					var innerErr error
					if v.State == Replaceable {
						//The release exists, it needs to be deleted
						dm := &cluster.DeleteMeta{
							ReleaseName:   v.ReleaseMeta.ReleaseName,
							Namespace:     v.ReleaseMeta.Namespace,
							DeleteTimeout: v.ReleaseMeta.InstallTimeout,
						}
						log.WithFields(log.Fields{
							"Name":        v.ReleaseMeta.ReleaseName,
							"Namespace":   v.ReleaseMeta.Namespace,
							"InstallWait": v.ReleaseMeta.InstallWait,
						}).Info("Deleting (force install)")
						if err := session.DeleteRelease(dm); err != nil {
							return errors.Wrap(err, "error deleting release before install (forced)")
						}
					}
					log.WithFields(log.Fields{
						"Name":        v.ReleaseMeta.ReleaseName,
						"Namespace":   v.ReleaseMeta.Namespace,
						"InstallWait": v.ReleaseMeta.InstallWait,
					}).Info("Installing")
					for i := 0; i < opt.InstallRetry; i++ {
						if control.Canceled() {
							return nil
						}
						msg, relName, err := session.InstallRelease(v.ReleaseMeta)
						if err != nil {
							log.WithFields(log.Fields{
								"Name":        v.ReleaseMeta.ReleaseName,
								"Namespace":   v.ReleaseMeta.Namespace,
								"InstallWait": v.ReleaseMeta.InstallWait,
								"Error":       err.Error(),
							}).Debug("Install reported error")
							innerErr = err

							if control.Canceled() {
								return nil
							}

							//The state has changed underneath us, but the release needs installed anyhow
							//So delete and try again
							dm := &cluster.DeleteMeta{
								ReleaseName:   v.ReleaseMeta.ReleaseName,
								Namespace:     v.ReleaseMeta.Namespace,
								DeleteTimeout: v.ReleaseMeta.InstallTimeout,
							}
							log.WithFields(log.Fields{
								"Name":        v.ReleaseMeta.ReleaseName,
								"Namespace":   v.ReleaseMeta.Namespace,
								"InstallWait": v.ReleaseMeta.InstallWait,
							}).Info("Deleting (state change)")
							if err := session.DeleteRelease(dm); err != nil {
								//deleting kube-proxy or other connection issues can trigger this, don't abort the retry
								log.Debug(err, "error deleting release before install (forced)")
							}
							if control.Canceled() {
								return nil
							}
							/////
							select {
							default:
								_ = <-time.After(1 * time.Second)
							}
							continue
						}
						log.WithFields(log.Fields{
							"Name":        v.ReleaseMeta.ReleaseName,
							"Namespace":   v.ReleaseMeta.Namespace,
							"InstallWait": v.ReleaseMeta.InstallWait,
							"Release":     relName,
						}).Info(msg)
						innerErr = nil
						return nil
					}
					return errors.WithFields(errors.Fields{
						"Name":        v.ReleaseMeta.ReleaseName,
						"Namespace":   v.ReleaseMeta.Namespace,
						"InstallWait": v.ReleaseMeta.InstallWait,
					}).Wrap(innerErr, "Error while installing release")
				}(); err != nil {
					control.Send(&ProcessResp{err: err})
					return
				}
			case Upgradable:
				if !v.Changed {
					log.WithFields(log.Fields{
						"Name":      v.ReleaseMeta.ReleaseName,
						"Namespace": v.ReleaseMeta.Namespace,
					}).Info("Skipping")
					return
				}
				log.WithFields(log.Fields{
					"Name":      v.ReleaseMeta.ReleaseName,
					"Namespace": v.ReleaseMeta.Namespace,
				}).Info("Upgrading")
				msg, err := session.UpgradeRelease(v.ReleaseMeta)
				if err != nil {
					control.Send(&ProcessResp{
						err: errors.WithFields(errors.Fields{
							"Name":      v.ReleaseMeta.ReleaseName,
							"Namespace": v.ReleaseMeta.Namespace,
						}).Wrap(err, "error while upgrading release"),
					})
					return
				}
				log.WithFields(log.Fields{
					"Name":      v.ReleaseMeta.ReleaseName,
					"Namespace": v.ReleaseMeta.Namespace,
				}).Info(msg)
			default:
				log.WithFields(log.Fields{
					"Name":      v.ReleaseMeta.ReleaseName,
					"Namespace": v.ReleaseMeta.Namespace,
				}).Info("Skipping")
			}
		}(control, releaseTarget)

		// sequenced is a user setting within a ChartGroup
		// true means we should only perform one action at a time within the ChartGroup
		if sequenced {
			// Wait for action to complete before moving to the next
			log.Debug("sequenced")
			if err := control.Wait(); err != nil {
				return err
			}
		} else {
			log.Debug("parallel apply")
		}
	}
	// make sure all actions are complete within this ReleaseGroup
	return control.Wait()
}

// NewControl returns a structure useful for flow control and messaging
func NewControl(inctx context.Context) *Control {
	ctx, ctxCancel := context.WithCancel(inctx)
	c := &Control{
		ctx:    ctx,
		cancel: ctxCancel,
		C:      make(chan *ProcessResp),
		wg:     sync.WaitGroup{},
	}
	return c
}

// Send a message to the controller instance
func (c *Control) Send(msg *ProcessResp) {
	c.C <- msg
}

// Add to the controllers sync.WaitGroup
func (c *Control) Add(n int) {
	c.wg.Add(n)
}

// Run Done() on the controllers sync.WaitGroup
// this should be done when the action returns
func (c *Control) Done() {
	c.wg.Done()
}

func (c *Control) Canceled() bool {
	select {
	case <-c.ctx.Done():
		return true
	default:
	}
	return false
}

// Wait blocks until all actions have concluded and potentially returns an error
func (c *Control) Wait() error {
	select {
	case resp := <-c.C:
		if resp.err != nil {
			// this can mask subsiquent errors since no one is checking the response channel
			log.Error(errors.Wrap(resp.err, "Got error from resp channel"))
			c.Cancel()
			c.wg.Wait()
			return resp.err
		}
	case <-func() chan struct{} {
		flowChan := make(chan struct{})
		go func() {
			c.wg.Wait()
			log.Info("processing completed")
			c.Cancel()
			flowChan <- struct{}{}
		}()
		return flowChan
	}():
	}
	return nil
}

// Cancel signals all actions to conclude now
func (c *Control) Cancel() {
	c.cancel()
}
