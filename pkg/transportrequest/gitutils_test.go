package transportrequest

import (
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/storer"
	"github.com/stretchr/testify/assert"
	"io"
	"testing"
)

type commitIteratorMock struct {
	commits []object.Commit
	index   int
}

func (iter *commitIteratorMock) Next() (*object.Commit, error) {
	i := iter.index
	iter.index++

	if i >= len(iter.commits) {
		return nil, io.EOF // real iterators also behave like this
	}
	return &iter.commits[i], nil
}

func (iter *commitIteratorMock) ForEach(cb func(c *object.Commit) error) error {
	for {
		c, err := iter.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		err = cb(c)
		if err == storer.ErrStop {
			break
		}
		if err != nil {
			return err
		}
	}

	return nil
}

func (iter *commitIteratorMock) Close() {

}

func TestRetrieveLabelStraightForward(t *testing.T) {

	commitIter := &commitIteratorMock{
		commits: []object.Commit{
			object.Commit{
				Hash:    plumbing.NewHash("1212121212121212121212121212121212121212"),
				Message: "this is a commit without TransportRequestId\n\nThis is the first and last line of the message body",
			},
			object.Commit{
				Hash:    plumbing.NewHash("3434343434343434343434343434343434343434"),
				Message: "this is a commit with TransportRequestId\n\nThis is the first line of the message body\nTransportRequest: 12345678",
			},
		},
	}
	labels, err := FindLabelsInCommits(commitIter, "TransportRequest")
	if assert.NoError(t, err) {
		assert.Equal(t, []string{"12345678"}, labels)
	}
}
