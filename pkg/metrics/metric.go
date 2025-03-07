package metrics

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/uptrace/bun"
	"go.uber.org/zap"

	"github.com/uptrace/uptrace/pkg/bunapp"
	"github.com/uptrace/uptrace/pkg/bunconv"
	"github.com/uptrace/uptrace/pkg/metrics/mql"
)

const (
	uptraceServiceGraphClientDuration = "uptrace.service_graph.client_duration"
	uptraceServiceGraphServerDuration = "uptrace.service_graph.server_duration"
	uptraceServiceGraphFailedRequests = "uptrace.service_graph.failed_requests"
)

type Metric struct {
	bun.BaseModel `bun:"metrics,alias:m"`

	ID        uint64 `json:"id,string" bun:",pk,autoincrement"`
	ProjectID uint32 `json:"projectId"`

	Name        string     `json:"name"`
	Description string     `json:"description"`
	Instrument  Instrument `json:"instrument"`
	Unit        string     `json:"unit" bun:",nullzero"`
	AttrKeys    []string   `json:"attrKeys" bun:",array"`

	CreatedAt time.Time `json:"createdAt" bun:",nullzero"`
	UpdatedAt time.Time `json:"updatedAt" bun:",nullzero"`

	// Payload
	NumTimeseries uint64 `json:"numTimeseries" bun:",scanonly"`
}

func SelectMetricMap(
	ctx context.Context, app *bunapp.App, projectID uint32,
) (map[string]*Metric, error) {
	metrics, err := SelectMetrics(ctx, app, projectID)
	if err != nil {
		return nil, err
	}

	m := make(map[string]*Metric, len(metrics))

	for _, metric := range metrics {
		m[metric.Name] = metric
	}

	return m, nil
}

func newDeletedMetric(projectID uint32, metricName string) *Metric {
	return &Metric{
		ProjectID:  projectID,
		Name:       metricName,
		Instrument: InstrumentDeleted,
	}
}

func SelectMetrics(ctx context.Context, app *bunapp.App, projectID uint32) ([]*Metric, error) {
	var metrics []*Metric
	if err := app.PG.NewSelect().
		Model(&metrics).
		Where("project_id = ?", projectID).
		OrderExpr("name ASC").
		Limit(10000).
		Scan(ctx); err != nil {
		return nil, err
	}
	return metrics, nil
}

func SelectMetric(ctx context.Context, app *bunapp.App, id uint64) (*Metric, error) {
	metric := new(Metric)
	if err := app.PG.NewSelect().
		Model(metric).
		Where("id = ?", id).
		Scan(ctx); err != nil {
		return nil, err
	}
	return metric, nil
}

func SelectMetricByName(
	ctx context.Context, app *bunapp.App, projectID uint32, name string,
) (*Metric, error) {
	metric := new(Metric)
	if err := app.PG.NewSelect().
		Model(metric).
		Where("name = ?", name).
		Where("project_id = ?", projectID).
		Limit(1).
		Scan(ctx); err != nil {
		return nil, err
	}
	return metric, nil
}

func UpsertMetric(ctx context.Context, app *bunapp.App, m *Metric) error {
	if m.CreatedAt.IsZero() {
		m.CreatedAt = time.Now()
	}
	if _, err := app.PG.NewInsert().
		Model(m).
		On("CONFLICT (project_id, name) DO UPDATE").
		Set("description = EXCLUDED.description").
		Set("unit = EXCLUDED.unit").
		Set("instrument = EXCLUDED.instrument").
		Set("attr_keys = EXCLUDED.attr_keys").
		Set("updated_at = EXCLUDED.updated_at").
		Exec(ctx); err != nil {
		return err
	}
	return nil
}

func CreateSystemMetrics(ctx context.Context, app *bunapp.App, projectID uint32) error {
	metrics := []Metric{
		{
			ProjectID:   projectID,
			Name:        uptraceServiceGraphClientDuration,
			Description: "Requests duration between two nodes as seen from the client",
			Instrument:  InstrumentSummary,
			Unit:        bunconv.UnitMicroseconds,
			AttrKeys: []string{
				"type",
				"client",
				"server",
				"deployment.environment",
				"service.namespace",
			},
		},
		{
			ProjectID:   projectID,
			Name:        uptraceServiceGraphServerDuration,
			Description: "Requests duration between two nodes as seen from the server",
			Instrument:  InstrumentSummary,
			Unit:        bunconv.UnitMicroseconds,
			AttrKeys: []string{
				"type",
				"client",
				"server",
				"deployment.environment",
				"service.namespace",
			},
		},
		{
			ProjectID:   projectID,
			Name:        uptraceServiceGraphFailedRequests,
			Description: "Total count of failed requests between two nodes",
			Instrument:  InstrumentCounter,
			AttrKeys: []string{
				"type",
				"client",
				"server",
				"deployment.environment",
				"service.namespace",
			},
		},
	}
	return UpsertMetrics(ctx, app, metrics)
}

func UpsertMetrics(ctx context.Context, app *bunapp.App, metrics []Metric) error {
	if _, err := app.PG.NewInsert().
		Model(&metrics).
		On("CONFLICT (project_id, name) DO UPDATE").
		Set("description = EXCLUDED.description").
		Set("unit = EXCLUDED.unit").
		Set("instrument = EXCLUDED.instrument").
		Set("attr_keys = EXCLUDED.attr_keys").
		Set("updated_at = now()").
		Returning("updated_at").
		Exec(ctx); err != nil {
		return err
	}

	seen := make(map[uint32]bool)
	for i := range metrics {
		metric := &metrics[i]

		if !metric.UpdatedAt.IsZero() || seen[metric.ProjectID] {
			continue
		}
		seen[metric.ProjectID] = true

		job := createDashboardsTask.NewJob(metric.ProjectID)
		job.OnceInPeriod(30 * time.Second)
		if err := app.MainQueue.AddJob(ctx, job); err != nil {
			app.Zap(ctx).Error("DefaultQueue.Add failed", zap.Error(err))
		}
	}

	return nil
}

//------------------------------------------------------------------------------

type MetricColumn struct {
	Unit  string `json:"unit" yaml:"unit,omitempty"`
	Color string `json:"color" yaml:"color,omitempty"`
}

func newMetricAlias(metric *Metric, alias string) mql.MetricAlias {
	return mql.MetricAlias{Name: metric.Name, Alias: alias}
}

func parseMetrics(ss []string) ([]mql.MetricAlias, error) {
	metrics := make([]mql.MetricAlias, len(ss))
	for i, s := range ss {
		metric, err := parseMetricAlias(s)
		if err != nil {
			return nil, err
		}
		metrics[i] = metric
	}
	return metrics, validateMetrics(metrics)
}

var aliasRE = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

func parseMetricAlias(s string) (mql.MetricAlias, error) {
	for _, sep := range []string{" as ", " AS "} {
		if ss := strings.Split(s, sep); len(ss) == 2 {
			name := strings.TrimSpace(ss[0])
			alias := strings.TrimSpace(ss[1])

			if !strings.HasPrefix(alias, "$") {
				return mql.MetricAlias{}, fmt.Errorf("alias %q must start with a dollar sign", alias)
			}
			alias = strings.TrimPrefix(alias, "$")

			if !aliasRE.MatchString(alias) {
				return mql.MetricAlias{}, fmt.Errorf("invalid alias: %q", alias)
			}

			return mql.MetricAlias{
				Name:  name,
				Alias: alias,
			}, nil
		}
	}
	return mql.MetricAlias{}, fmt.Errorf("can't parse metric alias %q", s)
}

func validateMetrics(metrics []mql.MetricAlias) error {
	seen := make(map[string]struct{}, len(metrics))
	for _, metric := range metrics {
		if metric.Name == "" {
			return fmt.Errorf("metric name is empty")
		}
		if metric.Alias == "" {
			return fmt.Errorf("metric alias is empty")
		}
		if _, ok := seen[metric.Alias]; ok {
			return fmt.Errorf("duplicated metric alias %q", metric.Alias)
		}
		seen[metric.Alias] = struct{}{}
	}
	return nil
}
