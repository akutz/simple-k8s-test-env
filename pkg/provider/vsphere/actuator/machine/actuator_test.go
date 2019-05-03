package machine_test

import (
	"testing"

	"vmware.io/sk8/pkg/provider"
	"vmware.io/sk8/pkg/provider/vsphere/config"
)

func TestNew(t *testing.T) {
	actuator := provider.NewMachineActuator(config.GroupName)
	if actuator == nil {
		t.FailNow()
	}
	t.Logf("%T", actuator)
}
