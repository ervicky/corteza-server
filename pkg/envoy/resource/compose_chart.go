package resource

import (
	"fmt"
	"strconv"

	"github.com/cortezaproject/corteza-server/compose/types"
)

type (
	// ComposeChart represents a ComposeChart
	ComposeChart struct {
		*base
		Res *types.Chart

		// Might keep track of related namespace
		RefNs   *Ref
		RefMods RefSet
	}
)

func NewComposeChart(res *types.Chart, nsRef string, mmRef []string) *ComposeChart {
	r := &ComposeChart{
		base:    &base{},
		RefMods: make(RefSet, len(mmRef)),
	}
	r.SetResourceType(COMPOSE_CHART_RESOURCE_TYPE)
	r.Res = res

	r.AddIdentifier(identifiers(res.Handle, res.Name, res.ID)...)

	r.RefNs = r.AddRef(COMPOSE_NAMESPACE_RESOURCE_TYPE, nsRef)
	for i, mRef := range mmRef {
		r.RefMods[i] = r.AddRef(COMPOSE_MODULE_RESOURCE_TYPE, mRef).Constraint(r.RefNs)
	}

	// Initial timestamps
	r.SetTimestamps(MakeCUDATimestamps(&res.CreatedAt, res.UpdatedAt, res.DeletedAt, nil))

	return r
}

func (r *ComposeChart) SysID() uint64 {
	return r.Res.ID
}

func (r *ComposeChart) Ref() string {
	return FirstOkString(r.Res.Handle, r.Res.Name, strconv.FormatUint(r.Res.ID, 10))
}

// FindComposeChart looks for the chart in the resources
func FindComposeChart(rr InterfaceSet, ii Identifiers) (ch *types.Chart) {
	var chRes *ComposeChart

	rr.Walk(func(r Interface) error {
		cr, ok := r.(*ComposeChart)
		if !ok {
			return nil
		}

		if cr.Identifiers().HasAny(ii) {
			chRes = cr
		}
		return nil
	})

	// Found it
	if chRes != nil {
		return chRes.Res
	}

	return nil
}

func ComposeChartErrUnresolved(ii Identifiers) error {
	return fmt.Errorf("compose chart unresolved %v", ii.StringSlice())
}
