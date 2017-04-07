package broker

import (
	"encoding/json"

	"github.com/fusor/ansible-service-broker/pkg/apb"
	logging "github.com/op/go-logging"
)

type ProvisionJob struct {
	spec          *apb.Spec
	parameters    *apb.Parameters
	clusterConfig apb.ClusterConfig
	log           *logging.Logger
}

type ProvisionMsg struct {
	JobToken string `json:"job_token"`
	SpecId   string `json:"spec_id"`
	Msg      string `json:"msg"`
}

func (m ProvisionMsg) Render() string {
	render, _ := json.Marshal(m)
	return string(render)
}

func NewProvisionJob(
	spec *apb.Spec, parameters *apb.Parameters,
	clusterConfig apb.ClusterConfig, log *logging.Logger,
) *ProvisionJob {
	return &ProvisionJob{spec: spec, parameters: parameters,
		clusterConfig: clusterConfig, log: log}
}

func (p *ProvisionJob) Run(token string, msgBuffer chan<- WorkMsg) {
	//p.emit(fmt.Sprintf("Provisioning %s\n", token))

	extCreds, err := apb.Provision(p.spec, p.parameters, p.clusterConfig, p.log)
	if err != nil {
		p.log.Error("broker::Provision error occurred.")
		p.log.Error("%s", err.Error())
		// send error message
		msgBuffer <- ProvisionMsg{token, p.spec.Id, err.Error()}
	}

	// send creds
	jsonmsg, _ := json.Marshal(extCreds)
	msgBuffer <- ProvisionMsg{token, p.spec.Id, string(jsonmsg)}
}
