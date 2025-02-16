package store

import (
	"context"
	"strconv"

	"github.com/cortezaproject/corteza-server/compose/types"
	"github.com/cortezaproject/corteza-server/pkg/envoy"
	"github.com/cortezaproject/corteza-server/pkg/envoy/resource"
	"github.com/cortezaproject/corteza-server/pkg/filter"
	"github.com/cortezaproject/corteza-server/store"
	stypes "github.com/cortezaproject/corteza-server/system/types"
)

type (
	composeNamespaceFilter types.NamespaceFilter
	composeModuleFilter    types.ModuleFilter
	composeRecordFilter    types.RecordFilter
	composePageFilter      types.PageFilter
	composeChartFilter     types.ChartFilter

	composeStore interface {
		store.ComposeAttachments
		store.ComposeCharts
		store.ComposeModuleFields
		store.ComposeModules
		store.ComposeNamespaces
		store.ComposePages
		store.ComposeRecordValues
		store.ComposeRecords
	}

	composeDecoder struct {
		resourceID []uint64
	}
)

func newComposeDecoder() *composeDecoder {
	return &composeDecoder{
		resourceID: make([]uint64, 0, 200),
	}
}

func (d *composeDecoder) decodeComposeNamespace(ctx context.Context, s composeStore, ff []*composeNamespaceFilter) *auxRsp {
	mm := make([]envoy.Marshaller, 0, 100)
	if ff == nil {
		return &auxRsp{
			mm: mm,
		}
	}

	var nn types.NamespaceSet
	var fn types.NamespaceFilter
	var err error
	for _, f := range ff {
		aux := *f

		if aux.Limit == 0 {
			aux.Limit = 1000
		}

		for {
			nn, fn, err = s.SearchComposeNamespaces(ctx, types.NamespaceFilter(aux))
			if err != nil {
				return &auxRsp{
					err: err,
				}
			}

			for _, n := range nn {
				d.resourceID = append(d.resourceID, n.ID)

				mm = append(mm, newComposeNamespace(n))
			}

			if f.NextPage != nil {
				aux.PageCursor = fn.NextPage
			} else {
				break
			}
		}
	}

	return &auxRsp{
		mm: mm,
	}
}

func (d *composeDecoder) decodeComposeModule(ctx context.Context, s composeStore, ff []*composeModuleFilter) *auxRsp {
	mm := make([]envoy.Marshaller, 0, 100)
	if ff == nil {
		return &auxRsp{
			mm: mm,
		}
	}

	var nn types.ModuleSet
	var fn types.ModuleFilter
	var err error
	for _, f := range ff {
		aux := *f

		if aux.Limit == 0 {
			aux.Limit = 1000
		}

		for {
			nn, fn, err = s.SearchComposeModules(ctx, types.ModuleFilter(aux))
			if err != nil {
				return &auxRsp{
					err: err,
				}
			}

			for _, n := range nn {
				d.resourceID = append(d.resourceID, n.ID)

				n.Fields, _, err = s.SearchComposeModuleFields(ctx, types.ModuleFieldFilter{
					ModuleID: []uint64{n.ID},
				})

				if err != nil {
					return &auxRsp{
						err: err,
					}
				}

				mm = append(mm, newComposeModule(n))
			}

			if f.NextPage != nil {
				aux.PageCursor = fn.NextPage
			} else {
				break
			}
		}
	}

	return &auxRsp{
		mm: mm,
	}
}

func (d *composeDecoder) decodeComposeRecord(ctx context.Context, s store.Storer, ff []*composeRecordFilter) *auxRsp {
	mm := make([]envoy.Marshaller, 0, 100)
	if ff == nil {
		return &auxRsp{
			mm: mm,
		}
	}

	// When decoding large amounts of records (milions) we can probably assume
	// that each system user exists somewhere in the record set.
	//
	// That said, it's probably cheeper to simply say that we probably need all of the
	// references to hold.
	// If a user is not there in the preparation step, we shouldn't fail straight away.
	//
	// @todo use some heuristic to determine when to just list all of the users
	//       and when to preprocess the records to find out what users to get.

	relUsers := make(resource.UserstampIndex)
	uu, _, err := store.SearchUsers(ctx, s, stypes.UserFilter{
		Paging: filter.Paging{
			Limit: 0,
		},
	})
	if err != nil {
		return &auxRsp{
			err: err,
		}
	}
	relUsers.Add(uu...)

	mapValues := func(r *types.Record) map[string]string {
		rr := make(map[string]string)
		for _, v := range r.Values {
			rr[v.Name] = v.Value
		}

		return rr
	}

	// Prepare a series of resource.ComposeRecord instances; one for each provided filter
	for _, f := range ff {
		aux := *f

		if aux.Limit == 0 {
			aux.Limit = 1000
		}

		mod, err := store.LookupComposeModuleByID(ctx, s, f.ModuleID)
		if err != nil {
			return &auxRsp{
				err: err,
			}
		}

		ff, _, err := store.SearchComposeModuleFields(ctx, s, types.ModuleFieldFilter{ModuleID: []uint64{mod.ID}})
		if err != nil {
			return &auxRsp{
				err: err,
			}
		}
		mod.Fields = ff

		// Refs
		auxRecord := &composeRecordAux{
			refMod:   strconv.FormatUint(f.ModuleID, 10),
			relMod:   mod,
			refNs:    strconv.FormatUint(f.NamespaceID, 10),
			relUsers: relUsers,
		}

		// Walker
		auxRecord.walker = func(cb func(r *resource.ComposeRecordRaw) error) error {
			var nn types.RecordSet
			var fn types.RecordFilter
			var err error

			for {
				nn, fn, err = s.SearchComposeRecords(ctx, mod, types.RecordFilter(aux))
				if err != nil {
					return err
				}

				for _, n := range nn {
					// Create a raw record
					r := &resource.ComposeRecordRaw{
						ID:     strconv.FormatUint(n.ID, 10),
						Values: mapValues(n),
						Ts:     resource.MakeCUDATimestamps(&n.CreatedAt, n.UpdatedAt, n.DeletedAt, nil),
						Us:     resource.MakeCUDOUserstamps(n.CreatedBy, n.UpdatedBy, n.DeletedBy, n.OwnedBy),
					}

					err = cb(r)
					if err != nil {
						return err
					}
				}

				if f.NextPage != nil {
					aux.PageCursor = fn.NextPage
				} else {
					break
				}
			}
			return nil
		}

		mm = append(mm, newComposeRecordFromAux(auxRecord))
	}

	return &auxRsp{
		mm: mm,
	}
}

func (d *composeDecoder) decodeComposePage(ctx context.Context, s composeStore, ff []*composePageFilter) *auxRsp {
	mm := make([]envoy.Marshaller, 0, 100)
	if ff == nil {
		return &auxRsp{
			mm: mm,
		}
	}

	var nn types.PageSet
	var fn types.PageFilter
	var err error
	for _, f := range ff {
		aux := *f

		if aux.Limit == 0 {
			aux.Limit = 1000
		}

		for {
			nn, fn, err = s.SearchComposePages(ctx, types.PageFilter(aux))
			if err != nil {
				return &auxRsp{
					err: err,
				}
			}

			for _, n := range nn {
				d.resourceID = append(d.resourceID, n.ID)

				mm = append(mm, newComposePage(n))
			}

			if fn.NextPage != nil {
				aux.PageCursor = fn.NextPage
			} else {
				break
			}
		}
	}

	return &auxRsp{
		mm: mm,
	}
}

func (d *composeDecoder) decodeComposeChart(ctx context.Context, s composeStore, ff []*composeChartFilter) *auxRsp {
	mm := make([]envoy.Marshaller, 0, 100)
	if ff == nil {
		return &auxRsp{
			mm: mm,
		}
	}

	var nn types.ChartSet
	var fn types.ChartFilter
	var err error
	for _, f := range ff {
		aux := *f

		if aux.Limit == 0 {
			aux.Limit = 1000
		}

		for {
			nn, fn, err = s.SearchComposeCharts(ctx, types.ChartFilter(aux))
			if err != nil {
				return &auxRsp{
					err: err,
				}
			}

			for _, n := range nn {
				d.resourceID = append(d.resourceID, n.ID)

				mm = append(mm, newComposeChart(n))
			}

			if f.Limit > 0 {
				break
			} else if fn.NextPage != nil {
				aux.PageCursor = fn.NextPage
			} else {
				break
			}
		}
	}

	return &auxRsp{
		mm: mm,
	}
}

// ComposeNamespace adds a new compose NamespaceFilter
func (df *DecodeFilter) ComposeNamespace(f *types.NamespaceFilter) *DecodeFilter {
	if df.composeNamespace == nil {
		df.composeNamespace = make([]*composeNamespaceFilter, 0, 1)
	}
	df.composeNamespace = append(df.composeNamespace, (*composeNamespaceFilter)(f))
	return df
}

// ComposeModule adds a new compose ModuleFilter
func (df *DecodeFilter) ComposeModule(f *types.ModuleFilter) *DecodeFilter {
	if df.composeModule == nil {
		df.composeModule = make([]*composeModuleFilter, 0, 1)
	}
	df.composeModule = append(df.composeModule, (*composeModuleFilter)(f))
	return df
}

// ComposeRecord adds a new compose RecordFilter
func (df *DecodeFilter) ComposeRecord(f *types.RecordFilter) *DecodeFilter {
	if df.composeRecord == nil {
		df.composeRecord = make([]*composeRecordFilter, 0, 1)
	}
	df.composeRecord = append(df.composeRecord, (*composeRecordFilter)(f))
	return df
}

// ComposePage adds a new compose PageFilter
func (df *DecodeFilter) ComposePage(f *types.PageFilter) *DecodeFilter {
	if df.composePage == nil {
		df.composePage = make([]*composePageFilter, 0, 1)
	}
	df.composePage = append(df.composePage, (*composePageFilter)(f))
	return df
}

// ComposeChart adds a new compose ChartFilter
func (df *DecodeFilter) ComposeChart(f *types.ChartFilter) *DecodeFilter {
	if df.composeChart == nil {
		df.composeChart = make([]*composeChartFilter, 0, 1)
	}
	df.composeChart = append(df.composeChart, (*composeChartFilter)(f))
	return df
}
