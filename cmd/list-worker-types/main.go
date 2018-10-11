package main

import (
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/taskcluster/slugid-go/slugid"
	tcclient "github.com/taskcluster/taskcluster-client-go"
	"github.com/taskcluster/taskcluster-client-go/tcawsprovisioner"
	"github.com/taskcluster/taskcluster-client-go/tcqueue"
)

type Queue tcqueue.Queue
type Provisioner tcawsprovisioner.AwsProvisioner

func main() {
	taskIDs := map[string]string{}
	myQueue := tcqueue.NewFromEnv()
	taskGroupID := slugid.Nice()
	created := time.Now()
	for _, wt := range AllWorkerTypes() {
		fmt.Println(wt)
		x := strings.Split(wt, "/")
		provisionerID := x[0]
		workerType := x[1]
		taskID := slugid.Nice()
		taskIDs[wt] = taskID
		payload := GenericWorkerPayload{
			MaxRunTime: 3600,
			Command: []string{
				`echo`,
			},
		}
		payloadJSON := mustCompileToRawMessage(payload)
		taskDef := &tcqueue.TaskDefinitionRequest{
			Created:      tcclient.Time(created),
			Deadline:     tcclient.Time(created.Add(time.Hour * 3)),
			Dependencies: []string{},
			Expires:      tcclient.Time(created.Add(time.Hour * 24 * 30)),
			Extra:        json.RawMessage("{}"),
			Metadata: struct {
				Description string `json:"description"`
				Name        string `json:"name"`
				Owner       string `json:"owner"`
				Source      string `json:"source"`
			}{
				Description: "Checking worker version on " + provisionerID + "/" + workerType,
				Name:        "Checking worker version on " + provisionerID + "/" + workerType,
				Owner:       "pmoore@mozilla.com",
				Source:      "https://github.com/petemoore",
			},
			Payload:       *payloadJSON,
			Priority:      "highest",
			ProvisionerID: provisionerID,
			Requires:      "all-completed",
			Retries:       5,
			Routes:        []string{},
			SchedulerID:   "-",
			Scopes:        []string{},
			Tags:          map[string]string{},
			TaskGroupID:   taskGroupID,
			WorkerType:    workerType,
		}
		tsr, err := myQueue.CreateTask(taskID, taskDef)
		fatalOnError(err)

		respJSON, err := json.MarshalIndent(tsr, "", "  ")
		fatalOnError(err)

		fmt.Println(string(respJSON))
	}

	fmt.Println("Task Group ID: " + taskGroupID)
}

func mustCompileToRawMessage(data interface{}) *json.RawMessage {
	bytes, err := json.Marshal(data)
	fatalOnError(err)
	var JSON json.RawMessage
	err = json.Unmarshal(bytes, &JSON)
	fatalOnError(err)
	return &JSON
}

func fatalOnError(err error) {
	if err != nil {
		log.Fatalf("Error:\n%v", err)
	}
}

func NewProvisioner() *Provisioner {
	p := tcawsprovisioner.NewFromEnv()
	P := Provisioner(*p)
	return &P
}

func NewQueue() *Queue {
	q := tcqueue.NewFromEnv()
	Q := Queue(*q)
	return &Q
}

func (p *Provisioner) AllWorkerTypes() []string {
	prov := tcawsprovisioner.AwsProvisioner(*p)
	wt, err := prov.ListWorkerTypes()
	if err != nil {
		panic(err)
	}
	return []string(*wt)
}

func AllWorkerTypes() []string {
	uniqueWorkerTypes := map[string]bool{}
	q := NewQueue()
	provisioners := q.AllProvisionerIDs()
	workerTypes := make([][]string, len(provisioners), len(provisioners))
	var wg sync.WaitGroup
	for i, p := range provisioners {
		if p == "test-provisioner" || p == "no-provisioning-nope" || p == "dummy-test-provisioner" || p == "test-dummy-provisioner" {
			continue
		}
		wg.Add(1)
		go func(p string, i int) {
			defer wg.Done()
			provWorkerTypes := q.ProvisionerWorkerTypes(p)
			workerTypes[i] = make([]string, len(provWorkerTypes), len(provWorkerTypes))
			for j, wt := range provWorkerTypes {
				workerTypes[i][j] = p + "/" + wt
			}
		}(p, i)
	}
	wg.Wait()
	for _, p := range workerTypes {
		for _, wt := range p {
			uniqueWorkerTypes[wt] = true
		}
	}

	// Now merge in known worker types according to AWS provisioner
	p := NewProvisioner()
	provisionerWorkerTypes := p.AllWorkerTypes()
	for _, wt := range provisionerWorkerTypes {
		uniqueWorkerTypes["aws-provisioner-v1/"+wt] = true
	}

	keys := make([]string, 0, len(uniqueWorkerTypes))
	for key := range uniqueWorkerTypes {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func (q *Queue) AllProvisionerIDs() []string {
	Q := tcqueue.Queue(*q)
	provisioners := []string{}
	var r *tcqueue.ListProvisionersResponse
	ct := ""
	for r == nil || r.ContinuationToken != "" {
		var err error
		r, err = (&Q).ListProvisioners(ct, "")
		if err != nil {
			panic(err)
		}
		ct = r.ContinuationToken
		for _, p := range r.Provisioners {
			provisioners = append(provisioners, p.ProvisionerID)
		}
	}
	return provisioners
}

func (q *Queue) ProvisionerWorkerTypes(provisionerID string) []string {
	Q := tcqueue.Queue(*q)
	workerTypes := []string{}
	var r *tcqueue.ListWorkerTypesResponse
	ct := ""
	for r == nil || r.ContinuationToken != "" {
		var err error
		r, err = (&Q).ListWorkerTypes(provisionerID, ct, "")
		if err != nil {
			panic(err)
		}
		ct = r.ContinuationToken
		for _, p := range r.WorkerTypes {
			workerTypes = append(workerTypes, p.WorkerType)
		}
	}
	return workerTypes
}
