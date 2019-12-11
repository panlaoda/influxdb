package main

import (
	"github.com/influxdata/influxdb"
	"github.com/influxdata/influxdb/chronograf"
	"github.com/influxdata/influxdb/pkger"
)

func Convert1To2Dashboard(d1 *chronograf.Dashboard) (pkger.Resource, error) {
	d2 := &influxdb.Dashboard{
		Name: d1.Name,
	}

	cvs := []interface{}{}

	for _, cell := range d1.Cells {
		c := influxdb.Cell{
			CellProperty: influxdb.CellProperty{
				X: cell.X,
				Y: cell.Y,
				W: cell.W,
				H: cell.H,
			},
		}

		v := influxdb.View{
			ViewContents: influxdb.ViewContents{
				Name: cell.Name,
			},
		}

		switch cell.Type {
		case "line":
			v.Properties = influxdb.XYViewProperties{
				Queries:    convertQueries(cell.Queries),
				Axes:       convertAxes(cell.Axes),
				Type:       "xy",
				Legend:     convertLegend(cell.Legend),
				Geom:       "line",
				ViewColors: convertColors(cell.CellColors),
				Note:       cell.Note,
			}
		case "line-stacked":
		case "line-stepplot":
		case "bar":
		case "line-plus-single-stat":
		case "single-stat":
		case "gauge":
		case "table":
		case "alerts":
		case "news":
		case "guide":
		case "note":
		default:
			v.Properties = influxdb.EmptyViewProperties{}
		}

		cvs = append(cvs, pkger.ConvertToCellView(c, v))
	}

	// TODO(desa): pass in cvs
	return pkger.DashboardToResource(*d2, nil, d1.Name), nil
}

func convertAxes(a map[string]chronograf.Axis) map[string]influxdb.Axis {
	m := map[string]influxdb.Axis{}
	for k, v := range a {
		m[k] = influxdb.Axis{
			Bounds: v.Bounds,
			Label:  v.Label,
			Prefix: v.Prefix,
			Suffix: v.Suffix,
			Base:   v.Base,
			Scale:  v.Scale,
		}
	}

	return m
}

func convertLegend(l chronograf.Legend) influxdb.Legend {
	return influxdb.Legend{
		Type:        l.Type,
		Orientation: l.Orientation,
	}
}

func convertColors(cs []chronograf.CellColor) []influxdb.ViewColor {
	vs := []influxdb.ViewColor{}
	for _, c := range cs {
		v := influxdb.ViewColor{
			ID:   c.ID,
			Type: c.Type,
			Hex:  c.Hex,
			Name: c.Name,
			// TODO(desa): need to turn into hex value
			//Value: c.Value,
		}
		vs = append(vs, v)
	}

	return vs
}

func convertQueries(qs []chronograf.DashboardQuery) []influxdb.DashboardQuery {
	ds := []influxdb.DashboardQuery{}
	for _, q := range qs {
		d := influxdb.DashboardQuery{
			// TODO(desa): possibly we should try to compile the query to flux that we can show the user.
			Text:     "//" + q.Command,
			EditMode: "advanced",
		}

		ds = append(ds, d)
	}

	return ds
}
