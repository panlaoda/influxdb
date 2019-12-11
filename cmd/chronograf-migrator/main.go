package main

import (
	"context"
	"fmt"

	"github.com/influxdata/influxdb/chronograf"
	"github.com/influxdata/influxdb/chronograf/bolt"
)

func main() {
	c := bolt.NewClient()
	c.Path = "/Users/michaeldesa/go/src/github.com/desa/chronograf-migrator/chronograf-v1.db"

	ctx := context.Background()

	if err := c.Open(ctx, nil, chronograf.BuildInfo{}); err != nil {
		panic(err)
	}

	dashboardStore := c.DashboardsStore

	ds, err := dashboardStore.All(ctx)
	if err != nil {
		panic(err)
	}

	for _, d := range ds {
		r, err := Convert1To2Dashboard(&d)
		if err != nil {
			panic(err)
		}
		fmt.Println(r)
	}
}
