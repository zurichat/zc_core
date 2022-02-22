package plugin

import (
	"context"
	"testing"

	"github.com/mitchellh/mapstructure"
	"go.mongodb.org/mongo-driver/bson"
)

type testService struct {
	store []*Plugin
}

func (t *testService) Create(ctx context.Context, p *Plugin) error {
	t.store = append(t.store, p)
	return nil
}

func (t *testService) FindOne(ctx context.Context, f interface{}) (*Plugin, error) {
	filter := f.(bson.M)
	pMap := make(map[string]interface{})
	dec, _ := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		TagName:    "json",
		Result:     &pMap,
		ZeroFields: true,
	})

	for _, v := range t.store {
		_ = dec.Decode(v)
		for k, val := range pMap {
			if filter[k] == val {
				return v, nil
			}
		}
	}

	return nil, Errorf(ENOENT, "record not found")
}

func (t *testService) FindMany(ctx context.Context, f interface{}) (ps []*Plugin, err error) {
	filter := f.(bson.M)
	pMap := make(map[string]interface{})
	dec, _ := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		TagName:    "json",
		Result:     &pMap,
		ZeroFields: true,
	})

	for _, v := range t.store {
		_ = dec.Decode(v)
		for k, val := range pMap {
			if filter[k] == val {
				ps = append(ps, v)
			}
		}
	}
	return
}

func (t *testService) Update(ctx context.Context, f interface{}, update Patch) error {
	p, _ := t.FindOne(ctx, f)
	result := make(map[string]interface{})
	dec, _ := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		TagName:    "json",
		Result:     &result,
		ZeroFields: true,
	})
	dec.Decode(update)
	udpMap := make(map[string]interface{})
	for k, v := range result {
		udpMap[k] = v
	}
	dec.Decode(p)
	pMap := make(map[string]interface{})
	for k, v := range result {
		pMap[k] = v
	}
	for k, v := range udpMap {
		pMap[k] = v
	}
	np := &Plugin{}
	mapstructure.Decode(pMap, &np)
	t.store[0] = np
	return nil
}

func (t *testService) Delete(ctx context.Context, f interface{}) error {
	return nil
}

func assertStatusCode(tb testing.TB, want, got int) {
	tb.Helper()
	if got != want {
		tb.Errorf("expected status code %d, but got %d", want, got)
	}
}

func assertStringsEqual(tb testing.TB, got, want string) {
	tb.Helper()
	if got != want {
		tb.Errorf("expected %q, but got %q", want, got)
	}
}
