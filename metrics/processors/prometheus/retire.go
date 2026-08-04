package prometheus

import (
	stdprom "github.com/prometheus/client_golang/prometheus"

	"github.com/InsideGallery/core/metrics"
)

// Series retirement (metrics.SeriesDeleter). A Prometheus child metric lives in its
// vec until the process exits, so a caller that stops feeding a series does not stop
// exporting it: every scrape keeps carrying the last value it ever recorded, frozen
// at whatever it was. For a label whose values come and go — a peer address, a pod
// IP, a tenant — that is cardinality growth with no ceiling, one dead series per
// subject that ever existed, and a frozen counter reads to an operator like an
// active fault rather than like history.
//
// Deleting the child is the only thing that removes the series, and this is where a
// caller reaches it: the vecs are private to the processor, so nothing outside this
// package can do it.

// vecDeleter is the delete half of *stdprom.MetricVec, which CounterVec, GaugeVec
// and HistogramVec all embed, so one sweep serves all three collector maps.
type vecDeleter interface {
	DeletePartialMatch(labels stdprom.Labels) int
}

// DeleteSeriesMatching deletes every child of name whose labels include all of tags,
// and forgets the memoized handles that resolved to those children.
//
// Tags become labels through exactly the path a record takes, so the labels matched
// here are the labels recording produced — a tag whose name needed sanitizing
// matches the sanitized label the series actually carries. The match is partial:
// labels the caller did not name may hold any value.
//
// Retirement is idempotent. A name that was never recorded, a match no child
// satisfies, and a second delete of the same subject are all silent successes:
// callers retire from lifecycle events, which run whether or not the subject ever
// recorded anything.
func (p *processor) DeleteSeriesMatching(name string, tags []string) error {
	match := labelsFromTags(tags)
	if len(match.names) == 0 {
		return metrics.ErrEmptySeriesMatch
	}

	normalized := sanitizeName(name)
	labels := matchLabels(match)

	// Children first, memo second, and the order is load-bearing. Forgetting first
	// would leave a window where a concurrent record resolves the child that is
	// about to be deleted and memoizes it again — an orphan in the memo for the life
	// of the process, absorbing records that reach no scrape. This way the worst
	// case is one record landing on an already-deleted child while the memo is being
	// rewritten, and the next record of that tuple re-creates the series.
	p.deleteChildren(normalized, labels)
	p.forgetHandles(normalized, labels)

	return nil
}

// deleteChildren deletes the matching children from every collector registered
// under name. The vec itself stays registered: a vec with no children exports
// nothing, and keeping it means a subject that comes back records into the
// collector it already had instead of re-registering one.
func (p *processor) deleteChildren(name string, labels stdprom.Labels) {
	p.mu.Lock()
	defer p.mu.Unlock()

	deleteVecChildren(p.counters, name, labels)
	deleteVecChildren(p.gauges, name, labels)
	deleteVecChildren(p.histograms, name, labels)
}

// deleteVecChildren sweeps one collector map. The map is keyed by (name, label
// keys), and Prometheus rejects two collectors that share a name with different
// label dimensions, so in practice one name has one entry — the loop is written to
// cover every entry under the name rather than to rely on that.
func deleteVecChildren[V vecDeleter](collectors map[collectorKey]V, name string, labels stdprom.Labels) {
	for key, collector := range collectors {
		if key.name != name {
			continue
		}

		collector.DeletePartialMatch(labels)
	}
}

// forgetHandles drops the memoized children the delete just invalidated. Without
// it a later record of a retired tuple would hit the memo, land on the deleted
// child, and export nothing — the series would never come back.
func (p *processor) forgetHandles(name string, labels stdprom.Labels) {
	matches := func(key handleKey) bool {
		return sanitizeName(key.name) == name && keyLabelsMatch(key, labels)
	}

	p.counterHandles.forget(matches)
	p.gaugeHandles.forget(matches)
	p.histogramHandles.forget(matches)
}

// keyLabelsMatch reports whether the labels a memoized tuple resolved to include all
// of labels. The key holds the caller's raw tags, so they are resolved rather than
// compared as strings: the memo has to forget exactly the children the vec sweep
// deleted, and both sides therefore decide through labelsFromTags. Unused key slots
// are empty strings, which carry no "name:value" and drop out of resolution.
func keyLabelsMatch(key handleKey, labels stdprom.Labels) bool {
	resolved := labelsFromTags(key.tags[:])

	for name, value := range labels {
		if !hasLabel(resolved, name, value) {
			return false
		}
	}

	return true
}

// hasLabel reports whether labels carries name with exactly value.
func hasLabel(labels labelSet, name, value string) bool {
	for index, candidate := range labels.names {
		if candidate == name {
			return labels.values[index] == value
		}
	}

	return false
}

// matchLabels turns a resolved label set into the partial match Prometheus takes.
func matchLabels(labels labelSet) stdprom.Labels {
	match := make(stdprom.Labels, len(labels.names))

	for index, name := range labels.names {
		match[name] = labels.values[index]
	}

	return match
}
