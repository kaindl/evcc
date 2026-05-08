package coordinator

import (
	"testing"

	"github.com/evcc-io/evcc/api"
	"github.com/evcc-io/evcc/core/loadpoint"
	"github.com/evcc-io/evcc/util"
	"go.uber.org/mock/gomock"
)

func TestVehicleDetectByStatus(t *testing.T) {
	ctrl := gomock.NewController(t)

	type vehicle struct {
		*api.MockVehicle
		*api.MockChargeState
	}

	v1 := &vehicle{api.NewMockVehicle(ctrl), api.NewMockChargeState(ctrl)}
	v2 := &vehicle{api.NewMockVehicle(ctrl), api.NewMockChargeState(ctrl)}

	type testcase struct {
		string
		v1, v2             api.ChargeStatus
		v2autodiscoverable bool
		res                api.Vehicle
	}
	tc := []testcase{
		{"A/A->0", api.StatusA, api.StatusA, true, nil},
		{"B/A->1", api.StatusB, api.StatusA, true, v1},
		{"B/A->1", api.StatusB, api.StatusA, true, v1},
		{"A/B->2", api.StatusA, api.StatusB, true, v2},
		{"A/B->2", api.StatusA, api.StatusB, true, v2},
		{"A/B->2", api.StatusA, api.StatusB, false, nil},
		{"A/B->2", api.StatusA, api.StatusB, false, nil},
		{"B/B->1", api.StatusB, api.StatusB, true, nil},
		{"B/B->1", api.StatusB, api.StatusB, false, v1},
	}

	log := util.NewLogger("foo")
	vehicles := []api.Vehicle{v1, v2}

	v1.MockVehicle.EXPECT().GetTitle().Return("v1").AnyTimes()
	v2.MockVehicle.EXPECT().GetTitle().Return("v2").AnyTimes()
	v1.MockVehicle.EXPECT().Identifiers().Return(nil).AnyTimes()
	v2.MockVehicle.EXPECT().Identifiers().Return([]string{"it's me"}).AnyTimes()
	v1.MockVehicle.EXPECT().AutoDiscoveryEnabled().Return(true).AnyTimes()

	var lp loadpoint.API
	c := New(log, vehicles)

	for _, tc := range tc {
		t.Logf("%+v", tc)

		v1.MockChargeState.EXPECT().Status().Return(tc.v1, nil)
		v2.MockVehicle.EXPECT().AutoDiscoveryEnabled().Return(tc.v2autodiscoverable)
		if tc.v2autodiscoverable {
			v2.MockChargeState.EXPECT().Status().Return(tc.v2, nil)
		}

		available := c.availableDetectibleVehicles(lp) // include id-able vehicles
		res := c.identifyVehicleByStatus(available)
		if tc.res != res {
			t.Errorf("expected %v, got %v", tc.res, res)
		}

		if res != nil {
			c.acquire(lp, res)
		} else {
			c.release(res)
		}
	}
}
