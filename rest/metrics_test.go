package rest_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-restclient/rest"
)

func TestCollector(t *testing.T) {
	actual := rest.HTTPCollector
	assert.NotNil(t, actual)
}

func BenchmarkCollectorInc(b *testing.B) {
	for i := 0; i < b.N; i++ {
		rest.HTTPCollector.IncrementCounter("client", "event_type", "sub_event_type")
	}
}

func BenchmarkCollectorRec(b *testing.B) {
	for i := 0; i < b.N; i++ {
		rest.HTTPCollector.RecordExecutionTime("client", "event_type", "sub_event_type", 1000)
	}
}
