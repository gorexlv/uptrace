package alerting

import (
	"bytes"
	"fmt"
	"html/template"
	"strings"

	"github.com/uptrace/uptrace/pkg/bunconv"
	"github.com/uptrace/uptrace/pkg/org"
	"github.com/uptrace/uptrace/pkg/utf8util"
)

var emailErrorFormatter = NewAlertFormatter(
	"<br />",
	WithCreatedTemplate(
		`🐞 [{{ $.projectName }}] A new error has just occurred`,
	),
	WithRecurringTemplate(
		`🐞 [{{ $.projectName }}] The error has {{ $.spanCount }} occurrences`,
	),
	WithClosedTemplate(
		`✅ [{{ $.projectName }}] The error is closed`,
	),
	WithReopenedTemplate(
		`🔴 [{{ $.projectName }}] The error is reopened`,
	),
	WithClosedByTemplate(
		`✅ [{{ $.projectName }}] The error is closed by {{ $.username }}`,
	),
	WithReopenedByTemplate(
		`🔴 [{{ $.projectName }}] The error is reopened by {{ $.username }}`,
	),
)

var telegramErrorFormatter = NewAlertFormatter(
	"\n",
	WithCreatedTemplate(
		`🐞 [{{ $.projectName }}] A new error has just occurred`,
		`{{ $.alertName }}`,
	),
	WithRecurringTemplate(
		`🐞 [{{ $.projectName }}] The error has {{ $.spanCount }} occurrences`,
		`{{ $.alertName }}`,
	),
	WithClosedTemplate(
		`✅ [{{ $.projectName }}] The error is closed`,
		`{{ $.alertName }}`,
	),
	WithReopenedTemplate(
		`🔴 [{{ $.projectName }}] The error is reopened`,
		`{{ $.alertName }}`,
	),
	WithClosedByTemplate(
		`✅ [{{ $.projectName }}] The error is closed by {{ $.username }}`,
		`{{ $.alertName }}`,
	),
	WithReopenedByTemplate(
		`🔴 [{{ $.projectName }}] The error is reopened by {{ $.username }}`,
		`{{ $.alertName }}`,
	),
)

var emailMetricFormatter = NewAlertFormatter(
	"<br />",
	WithCreatedTemplate(
		`🔥 Firing: {{ $.shortSummary }}`,
	),
	WithRecurringTemplate(
		`🔥 Firing for {{ $.duration }}: {{ $.shortSummary }}`,
	),
	WithClosedTemplate(
		`✅ Back to normal after {{ $.duration }}: {{ $.normalValue }} (was {{ $.shortSummary }})`,
	),
	WithReopenedTemplate(
		`🔴 Firing again: {{ $.shortSummary }}`,
	),
	WithClosedByTemplate(
		`✅ Closed by {{ $.username }}: {{ $.shortSummary }}`,
	),
	WithReopenedByTemplate(
		`🔴 Reopened by {{ $.username }}: {{ $.shortSummary }}`,
	),
)

var telegramMetricFormatter = NewAlertFormatter(
	"\n",
	WithCreatedTemplate(
		``,
		`🔥 Firing: {{ $.shortSummary }}`,
		`[{{ $.projectName }}] {{ $.alertName }}`,
		`{{ $.longSummary }}`,
	),
	WithRecurringTemplate(
		`🔥 Firing for {{ $.duration }}: {{ $.shortSummary }}`,
		`[{{ $.projectName }}] {{ $.alertName }}`,
		`{{ $.longSummary }}`,
	),
	WithClosedTemplate(
		`✅ Back to normal after {{ $.duration }}: {{ $.normalValue }} (was {{ $.shortSummary }})`,
		`[{{ $.projectName }}] {{ $.alertName }}`,
	),
	WithReopenedTemplate(
		`🔴 Firing again: {{ $.shortSummary }}`,
		`[{{ $.projectName }}] {{ $.alertName }}`,
	),
	WithClosedByTemplate(
		`✅ Closed by {{ $.username }}: {{ $.shortSummary }}`,
		`[{{ $.projectName }}] {{ $.alertName }}`,
	),
	WithReopenedByTemplate(
		`🔴 Reopened by {{ $.username }}: {{ $.shortSummary }}`,
		`[{{ $.projectName }}] {{ $.alertName }}`,
	),
)

type AlertFormatter struct {
	breakLine  string
	created    *template.Template
	recurring  *template.Template
	closed     *template.Template
	reopened   *template.Template
	closedBy   *template.Template
	reopenedBy *template.Template
}

type AlertFormatterOption func(*AlertFormatter)

func WithCreatedTemplate(tpl ...string) AlertFormatterOption {
	return func(f *AlertFormatter) {
		f.created = f.newTemplate(tpl)
	}
}

func WithRecurringTemplate(tpl ...string) AlertFormatterOption {
	return func(f *AlertFormatter) {
		f.recurring = f.newTemplate(tpl)
	}
}

func WithClosedTemplate(tpl ...string) AlertFormatterOption {
	return func(f *AlertFormatter) {
		f.closed = f.newTemplate(tpl)
	}
}

func WithReopenedTemplate(tpl ...string) AlertFormatterOption {
	return func(f *AlertFormatter) {
		f.reopened = f.newTemplate(tpl)
	}
}

func WithClosedByTemplate(tpl ...string) AlertFormatterOption {
	return func(f *AlertFormatter) {
		f.closedBy = f.newTemplate(tpl)
	}
}

func WithReopenedByTemplate(tpl ...string) AlertFormatterOption {
	return func(f *AlertFormatter) {
		f.reopenedBy = f.newTemplate(tpl)
	}
}

func NewAlertFormatter(breakLine string, opts ...AlertFormatterOption) *AlertFormatter {
	f := &AlertFormatter{
		breakLine: breakLine,
	}
	for _, opt := range opts {
		opt(f)
	}
	return f
}

func (f *AlertFormatter) Format(project *org.Project, alert org.Alert) string {
	baseAlert := alert.Base()
	params := map[string]any{
		"projectName": project.Name,
		"alert":       alert,
		"alertName":   utf8util.Trunc(baseAlert.Name, 255),
	}

	if baseAlert.Event.User != nil {
		params["username"] = baseAlert.Event.User.Username()
	}

	switch alert := alert.(type) {
	case *ErrorAlert:
		params["spanCount"] = alert.Params.SpanCount
	case *MetricAlert:
		params["shortSummary"] = alert.ShortSummary()
		params["longSummary"] = alert.LongSummary(f.breakLine)
		params["duration"] = bunconv.ShortDuration(alert.Event.CreatedAt.Sub(alert.Event.Time))

		if alert.Event.Status == org.AlertStatusClosed {
			unit := alert.Params.Monitor.ColumnUnit
			params["normalValue"] = bunconv.Format(alert.Params.NormalValue, unit)
		}
	}

	switch baseAlert.Event.Name {
	case org.AlertEventCreated:
		return f.format(f.created, params)
	case org.AlertEventRecurring:
		return f.format(f.recurring, params)
	case org.AlertEventStatusChanged:
		switch baseAlert.Event.Status {
		case org.AlertStatusOpen:
			if baseAlert.Event.User != nil {
				return f.format(f.reopenedBy, params)
			}
			return f.format(f.reopened, params)
		case org.AlertStatusClosed:
			if baseAlert.Event.User != nil {
				return f.format(f.closedBy, params)
			}
			return f.format(f.closed, params)
		default:
			return fmt.Sprintf("unsupported alert status: %q", baseAlert.Event.Status)
		}
	default:
		return fmt.Sprintf("unsupported alert event: %q", baseAlert.Event.Name)
	}
}

func (f *AlertFormatter) newTemplate(tpl []string) *template.Template {
	return template.Must(template.New("").Parse(strings.Join(tpl, "\n")))
}

func (f *AlertFormatter) format(tpl *template.Template, params map[string]any) string {
	if tpl == nil {
		return ""
	}
	var buf bytes.Buffer
	if err := tpl.Execute(&buf, params); err != nil {
		return err.Error()
	}
	return buf.String()
}
