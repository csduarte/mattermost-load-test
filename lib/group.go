package lib

import (
	"fmt"
	"time"
)

// Group is a simple container for the threads and stats aggregation
type Group struct {
	Total        int
	LaunchCount  int
	ActiveCount  int
	ActionCount  int
	Errors       []string
	ActivityPipe chan Activity
}

func (g *Group) initialize() {
	g.Errors = []string{}
	g.ActivityPipe = make(chan Activity, 40000)
}

// Kickstart will kick off group and connect channel listeners
func (g *Group) Kickstart(tpGen TestPlanGen, total, offset, SecRamp int) {

	g.initialize()
	g.Total = total

	// Generate a Test Plan for Global Setup
	testPlan := tpGen(0, nil)
	err := testPlan.GlobalSetup()

	if err != nil {
		panic(err)
	}

	sleepIncrement := time.Duration(SecRamp) * time.Second / time.Duration(total)
	go g.spinUpThreads(tpGen, total, offset, sleepIncrement)

	for activity := range g.ActivityPipe {
		switch activity.Status {
		case StatusActive:
			g.registerThreadActive(activity)
		case StatusInactive:
			g.registerThreadInactive(activity)
		case StatusLaunching:
			g.registerThreadLaunching(activity)
		case StatusLaunchFailed:
			g.registerLaunchFail(activity)
		case StatusError:
			g.registerThreadError(activity)
		case StatusAction:
			g.registerThreadAction(activity)
		default:
			panic("Unhandled Activity type in group")
		}
	}
}

func (g *Group) spinUpThreads(tp TestPlanGen, total, start int, sleep time.Duration) {
	for i := start; i < total+start; i++ {
		t := Thread{id: i}
		go t.Start(tp, g.ActivityPipe)
		time.Sleep(sleep)
	}
}

func (g *Group) registerThreadLaunching(activity Activity) {
	g.LaunchCount++
}

func (g *Group) registerLaunchFail(activity Activity) {
	g.LaunchCount--
}

func (g *Group) registerThreadFinished(activity Activity) {
	g.LaunchCount--
}

func (g *Group) registerThreadAction(activity Activity) {
	g.ActionCount++
}

func (g *Group) registerThreadActive(activity Activity) {
	g.LaunchCount--
	g.ActiveCount++
}

func (g *Group) registerThreadInactive(activity Activity) {
	g.ActiveCount--
}

func (g *Group) registerThreadError(activity Activity) {
	errMsg := fmt.Sprintf("Thread #%d - %v - %v", activity.ID, activity.Message, activity.Err.Error())
	g.Errors = append(g.Errors, errMsg)
}
