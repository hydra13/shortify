package resettablepoll

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type RstMock struct {
	Name string
}

func (r *RstMock) Reset() {
	r.Name = ""
}

func TestPoll(t *testing.T) {
	t.Parallel()

	p := New[*RstMock]()

	assert.Equal(t, 0, len(p.items))

	p.Put(&RstMock{Name: "Test1"})
	p.Put(&RstMock{Name: "Test2"})
	p.Put(&RstMock{Name: "Test3"})

	assert.Equal(t, 3, len(p.items))

	item := p.Get()
	assert.NotNil(t, item)
	assert.Empty(t, item.(*RstMock).Name)

	for range 3 {
		item = p.Get()
	}

	assert.Nil(t, item)
}
